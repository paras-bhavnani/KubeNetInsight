from sentence_transformers import SentenceTransformer
import numpy as np

class EmbeddingModel:
    def __init__(self, model_name='all-MiniLM-L6-v2'):
        self.model = SentenceTransformer(model_name)
        
    def generate_embeddings(self, texts):
        return self.model.encode(texts)
    
    def evaluate_performance(self, test_texts):
        import time
        start_time = time.time()
        embeddings = self.generate_embeddings(test_texts)
        end_time = time.time()
        
        return {
            'time_taken': end_time - start_time,
            'embedding_dimension': embeddings.shape[1],
            'memory_usage': embeddings.nbytes / 1024  # KB
        }

if __name__ == "__main__":
    # Test the implementation
    model = EmbeddingModel()
    test_texts = [
        "Troubleshooting pod creation failures in Kubernetes",
        "Network connectivity issues in KubeNetInsight",
        "Performance optimization for container deployments"
    ]
    
    print("Testing embedding generation...")
    results = model.evaluate_performance(test_texts)
    print("Results:", results)