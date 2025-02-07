import numpy as np
import sys
import os

# Add the project root to Python path for imports
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(__file__))))

from pkg.embeddings.index import DocumentStore

class QueryOptimizer:
    def __init__(self, document_store):
        self.document_store = document_store

    def optimize_query(self, query, k=5):
        """Optimize the query by normalizing distances and ranking results."""
        results = self.document_store.search(query, k)
        
        # Normalize distances to a 0-1 scale
        distances = np.array([result['distance'] for result in results])
        max_distance = distances.max()
        min_distance = distances.min()
        normalized_distances = (distances - min_distance) / (max_distance - min_distance + 1e-9)

        # Update results with normalized distances
        for i, result in enumerate(results):
            result['normalized_distance'] = normalized_distances[i]

        # Sort results by normalized distance (ascending)
        results = sorted(results, key=lambda x: x['normalized_distance'])
        return results

if __name__ == "__main__":
    # Test the QueryOptimizer
    doc_store = DocumentStore()
    doc_store.load("kubeNetInsight_index.faiss", "kubeNetInsight_docs.npy")

    optimizer = QueryOptimizer(doc_store)

    test_queries = [
        "How to fix pod failures",
        "Network connectivity troubleshooting",
        "Performance optimization guide",
        "Prometheus metrics setup"
    ]

    for query in test_queries:
        print(f"\nQuery: '{query}'")
        optimized_results = optimizer.optimize_query(query, k=3)
        print("Top 3 optimized results:")
        for i, result in enumerate(optimized_results, 1):
            doc_preview = result['document'][:100] + "..." if len(result['document']) > 100 else result['document']
            print(f"{i}. {doc_preview} (normalized distance: {result['normalized_distance']:.2f})")
