import os
import json
import logging
from datetime import datetime
from flask import Flask, request, jsonify
import pandas as pd
import numpy as np
from sklearn.ensemble import IsolationForest
from sklearn.preprocessing import StandardScaler
import joblib
import psycopg2
from psycopg2.extras import RealDictCursor

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = Flask(__name__)

DB_CONFIG = {
    'host': os.getenv('DB_HOST', 'localhost'),
    'port': os.getenv('DB_PORT', '5432'),
    'database': os.getenv('DB_NAME', 'payments'),
    'user': os.getenv('DB_USER', 'postgres'),
    'password': os.getenv('DB_PASSWORD', 'postgres')
}

MODEL_PATH = 'models/fraud_model.pkl'
SCALER_PATH = 'models/scaler.pkl'

model = None
scaler = None


def get_db_connection():
    return psycopg2.connect(**DB_CONFIG)


def extract_features(transaction_data):
    features = {}

    features['amount'] = float(transaction_data.get('amount', 0))

    timestamp = datetime.fromisoformat(transaction_data.get('timestamp', datetime.now().isoformat()))
    features['hour'] = timestamp.hour
    features['day_of_week'] = timestamp.weekday()

    features['velocity_5min'] = transaction_data.get('velocity_5min', 0)
    features['velocity_1hour'] = transaction_data.get('velocity_1hour', 0)

    features['avg_amount_24h'] = transaction_data.get('avg_amount_24h', features['amount'])
    features['std_amount_24h'] = transaction_data.get('std_amount_24h', 0)

    if features['std_amount_24h'] > 0:
        features['amount_zscore'] = (features['amount'] - features['avg_amount_24h']) / features['std_amount_24h']
    else:
        features['amount_zscore'] = 0

    features['time_since_last_tx'] = transaction_data.get('time_since_last_tx', 3600)

    return features


def load_training_data():
    try:
        conn = get_db_connection()
        cursor = conn.cursor(cursor_factory=RealDictCursor)

        query = """
        SELECT
            t.amount,
            EXTRACT(HOUR FROM t.created_at) as hour,
            EXTRACT(DOW FROM t.created_at) as day_of_week,
            fs.velocity_score,
            fs.amount_score,
            fs.time_score,
            fs.pattern_score,
            fs.total_score,
            CASE WHEN fa.id IS NOT NULL THEN 1 ELSE 0 END as is_fraud
        FROM transactions t
        LEFT JOIN fraud_scores fs ON t.id = fs.transaction_id
        LEFT JOIN fraud_alerts fa ON t.id = fa.transaction_id AND fa.status = 'CONFIRMED'
        WHERE t.status = 'COMPLETED'
        ORDER BY t.created_at DESC
        LIMIT 10000
        """

        cursor.execute(query)
        data = cursor.fetchall()

        cursor.close()
        conn.close()

        if len(data) < 100:
            logger.warning(f"Insufficient training data: {len(data)} samples. Need at least 100.")
            return None

        df = pd.DataFrame(data)
        logger.info(f"Loaded {len(df)} transactions for training")

        return df

    except Exception as e:
        logger.error(f"Error loading training data: {e}")
        return None


def train_model():
    global model, scaler

    logger.info("Starting model training...")

    df = load_training_data()
    if df is None or len(df) < 100:
        logger.error("Cannot train model: insufficient data")
        return False

    feature_cols = ['amount', 'hour', 'day_of_week', 'velocity_score',
                    'amount_score', 'time_score', 'pattern_score']

    X = df[feature_cols].fillna(0)

    scaler = StandardScaler()
    X_scaled = scaler.fit_transform(X)

    fraud_rate = df['is_fraud'].mean() if 'is_fraud' in df else 0.01
    contamination = max(0.01, min(0.5, fraud_rate))

    model = IsolationForest(
        contamination=contamination,
        random_state=42,
        n_estimators=100,
        max_samples='auto',
        max_features=1.0
    )

    model.fit(X_scaled)

    os.makedirs('models', exist_ok=True)
    joblib.dump(model, MODEL_PATH)
    joblib.dump(scaler, SCALER_PATH)

    logger.info(f"Model trained successfully with contamination={contamination:.4f}")
    logger.info(f"Training samples: {len(X)}, Features: {len(feature_cols)}")

    return True


def load_model():
    global model, scaler

    try:
        if os.path.exists(MODEL_PATH) and os.path.exists(SCALER_PATH):
            model = joblib.load(MODEL_PATH)
            scaler = joblib.load(SCALER_PATH)
            logger.info("Model and scaler loaded successfully")
            return True
        else:
            logger.warning("Model files not found. Training new model...")
            return train_model()
    except Exception as e:
        logger.error(f"Error loading model: {e}")
        return False


def predict_fraud(features):
    global model, scaler

    if model is None or scaler is None:
        logger.warning("Model not loaded, loading now...")
        if not load_model():
            return {'ml_score': 0, 'is_anomaly': False, 'anomaly_score': 0}

    feature_cols = ['amount', 'hour', 'day_of_week', 'velocity_score',
                    'amount_score', 'time_score', 'pattern_score']

    feature_values = [features.get(col, 0) for col in feature_cols]
    X = np.array([feature_values])

    X_scaled = scaler.transform(X)

    prediction = model.predict(X_scaled)[0]
    anomaly_score = model.decision_function(X_scaled)[0]

    ml_score = int(max(0, min(100, (-anomaly_score + 0.5) * 50)))

    is_anomaly = prediction == -1

    return {
        'ml_score': ml_score,
        'is_anomaly': is_anomaly,
        'anomaly_score': float(anomaly_score)
    }


@app.route('/health', methods=['GET'])
def health():
    return jsonify({
        'status': 'healthy',
        'model_loaded': model is not None,
        'scaler_loaded': scaler is not None,
        'timestamp': datetime.now().isoformat()
    })


@app.route('/predict', methods=['POST'])
def predict():
    try:
        data = request.get_json()

        features = extract_features(data)

        features['amount_score'] = data.get('amount_score', 0)
        features['velocity_score'] = data.get('velocity_score', 0)
        features['time_score'] = data.get('time_score', 0)
        features['pattern_score'] = data.get('pattern_score', 0)

        result = predict_fraud(features)

        return jsonify({
            'success': True,
            'ml_score': result['ml_score'],
            'is_anomaly': result['is_anomaly'],
            'anomaly_score': result['anomaly_score'],
            'features_used': features
        })

    except Exception as e:
        logger.error(f"Prediction error: {e}")
        return jsonify({
            'success': False,
            'error': str(e),
            'ml_score': 0
        }), 500


@app.route('/train', methods=['POST'])
def train():
    try:
        success = train_model()

        if success:
            return jsonify({
                'success': True,
                'message': 'Model trained successfully',
                'timestamp': datetime.now().isoformat()
            })
        else:
            return jsonify({
                'success': False,
                'message': 'Training failed'
            }), 500

    except Exception as e:
        logger.error(f"Training error: {e}")
        return jsonify({
            'success': False,
            'error': str(e)
        }), 500


@app.route('/model/info', methods=['GET'])
def model_info():
    global model, scaler

    info = {
        'model_loaded': model is not None,
        'scaler_loaded': scaler is not None,
        'model_type': 'IsolationForest',
    }

    if model is not None:
        info['n_estimators'] = model.n_estimators
        info['contamination'] = model.contamination
        info['max_samples'] = model.max_samples

    return jsonify(info)


if __name__ == '__main__':
    logger.info("Starting ML Service...")
    load_model()

    port = int(os.getenv('PORT', 5000))
    app.run(host='0.0.0.0', port=port, debug=False)
