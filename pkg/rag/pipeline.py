from pkg.embeddings.index import get_context
from pkg.llm.inference import LlamaModel

class RAGPipeline:
    def __init__(self, model_path):
        self.llm = LlamaModel(model_path)

    def process_query(self, query):
        context = get_context(query)  # Retrieve context from FAISS
        prompt = f"Context: {context}\n\nQuestion: {query}\n\nAnswer:"
        return self.llm.generate_response(prompt, max_new_tokens=200)

if __name__ == "__main__":
    # Test RAG pipeline
    # model_path = "pkg/llm/models/Llama-2-7b"
    # Use the converted or pre-converted Hugging Face-compatible model path
    model_path = "meta-llama/Llama-2-7b-hf" # from hugginface
    rag_pipeline = RAGPipeline(model_path)
    print(rag_pipeline.process_query("How do I troubleshoot pod crashes?"))
