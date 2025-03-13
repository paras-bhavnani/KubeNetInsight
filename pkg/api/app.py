from flask import Flask, request, jsonify
from pkg.rag.pipeline import RAGPipeline

app = Flask(__name__)
# rag_pipeline = RAGPipeline("pkg/llm/models/Llama-2-7b")
rag_pipeline = RAGPipeline("meta-llama/Llama-2-7b-hf0")

@app.route("/rag", methods=["GET"])
def rag_endpoint():
    query = request.args.get("query")
    if not query:
        return jsonify({"error": "Query parameter is required"}), 400

    response = rag_pipeline.process_query(query)
    return jsonify({"response": response})

if __name__ == "__main__":
    # Test API locally
    app.run(host="0.0.0.0", port=8000)
