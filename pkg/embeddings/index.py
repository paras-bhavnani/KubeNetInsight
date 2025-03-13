import faiss
import numpy as np

import sys
import os
# Add the project root to Python path for imports
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(__file__))))

from pkg.embeddings.model import EmbeddingModel

def get_context(query, k=3):
    """Retrieve the top-k most relevant documents for a given query."""
    # Initialize DocumentStore
    doc_store = DocumentStore()
    
    # Load pre-saved FAISS index and documents
    index_path = "kubeNetInsight_index.faiss"
    documents_path = "kubeNetInsight_docs.npy"
    
    if not os.path.exists(index_path) or not os.path.exists(documents_path):
        raise FileNotFoundError("FAISS index or documents file not found. Please ensure they are saved.")
    
    doc_store.load(index_path, documents_path)
    
    # Search for relevant documents
    results = doc_store.search(query, k=k)
    
    # Combine the top-k results into a single context string
    context = "\n".join([result['document'] for result in results])
    return context

class DocumentStore:
    def __init__(self, dimension=384):
        self.dimension = dimension
        self.index = faiss.IndexFlatL2(dimension)
        self.embedding_model = EmbeddingModel()
        self.documents = []
        
    def add_documents(self, texts):
        """Add documents and their embeddings to the index."""
        embeddings = self.embedding_model.generate_embeddings(texts)
        self.index.add(embeddings.astype('float32'))
        self.documents.extend(texts)
        
    def search(self, query, k=5):
        """Search for similar documents using a query string."""
        query_embedding = self.embedding_model.generate_embeddings([query])
        distances, indices = self.index.search(query_embedding.astype('float32'), k)
        
        results = []
        for i, idx in enumerate(indices[0]):
            results.append({
                'document': self.documents[idx],
                'distance': distances[0][i]
            })
        return results
    
    def save(self, index_path, documents_path):
        """Save the FAISS index and documents."""
        faiss.write_index(self.index, index_path)
        np.save(documents_path, np.array(self.documents))
        
    def load(self, index_path, documents_path):
        """Load the FAISS index and documents."""
        self.index = faiss.read_index(index_path)
        self.documents = np.load(documents_path, allow_pickle=True).tolist()

if __name__ == "__main__":
    import os
    import glob
    
    doc_store = DocumentStore()
    
    # Load actual runbook content
    runbook_path = "manifests/documentation/runbooks/*.md"
    documents = []
    
    print("Loading runbook content...")
    for file_path in glob.glob(runbook_path):
        with open(file_path, 'r') as f:
            content = f.read()
            # Split content into meaningful chunks
            sections = content.split('\n## ')
            # Add each section as a separate document
            for section in sections:
                if section.strip():
                    documents.append(section.strip())
    
    print(f"Found {len(documents)} document sections")
    
    print("\nAdding documents to the store...")
    doc_store.add_documents(documents)
    
    print("\nTesting search functionality...")
    test_queries = [
        "How to fix pod failures",
        "Network connectivity troubleshooting",
        "Performance optimization guide",
        "Prometheus metrics setup"
    ]
    
    for query in test_queries:
        print(f"\nQuery: '{query}'")
        results = doc_store.search(query, k=3)
        print("Top 3 results:")
        for i, result in enumerate(results, 1):
            # Truncate long documents for display
            doc_preview = result['document'][:100] + "..." if len(result['document']) > 100 else result['document']
            print(f"{i}. {doc_preview} (distance: {result['distance']:.2f})")
    
    print("\nTesting save/load functionality...")
    doc_store.save("kubeNetInsight_index.faiss", "kubeNetInsight_docs.npy")
