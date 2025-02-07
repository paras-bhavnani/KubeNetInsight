import os
import sys
from flask import Flask, request, jsonify
import numpy as np

# Add project root to Python path
current_dir = os.path.dirname(os.path.abspath(__file__))
project_root = os.path.dirname(os.path.dirname(current_dir))
sys.path.append(project_root)

from pkg.embeddings.index import DocumentStore
from pkg.embeddings.optimizer import QueryOptimizer

app = Flask(__name__)

def make_json_serializable(results):
    for result in results:
        # Convert float32 to float for JSON serialization
        result['distance'] = float(result['distance'])
        if 'normalized_distance' in result:
            result['normalized_distance'] = float(result['normalized_distance'])
    return results

# Initialize document store and optimizer
doc_store = DocumentStore()
doc_store.load("kubeNetInsight_index.faiss", "kubeNetInsight_docs.npy")
optimizer = QueryOptimizer(doc_store)

@app.route('/', methods=['GET'])
def home():
    return jsonify({
        "status": "running",
        "endpoints": {
            "search": {
                "url": "/search",
                "method": "POST",
                "parameters": {
                    "query": "string",
                    "k": "integer (optional, default=5)"
                }
            }
        }
    })

@app.route('/search', methods=['POST'])
def search():
    data = request.get_json()
    query = data.get('query', '')
    k = data.get('k', 5)

    if not query:
        return jsonify({"error": "Query is required"}), 400

    results = optimizer.optimize_query(query, k)
    serializable_results = make_json_serializable(results)
    return jsonify(serializable_results)

if __name__ == "__main__":
    app.run(debug=True)
