from transformers import AutoTokenizer, AutoModelForCausalLM

class LlamaModel:
    def __init__(self, model_path):
        self.tokenizer = AutoTokenizer.from_pretrained(model_path)
        self.model = AutoModelForCausalLM.from_pretrained(model_path, device_map="auto")

    def generate_response(self, prompt, max_new_tokens=100):
        inputs = self.tokenizer(prompt, return_tensors="pt").to(self.model.device)
        outputs = self.model.generate(**inputs, max_new_tokens=max_new_tokens)
        return self.tokenizer.decode(outputs[0], skip_special_tokens=True)

if __name__ == "__main__":
    # Test the model
    # model_path = "pkg/llm/models/Llama-2-7b" # model from meta
    # Use the converted or pre-converted Hugging Face-compatible model path
    model_path = "meta-llama/Llama-2-7b-hf" # from hugginface
    llama = LlamaModel(model_path)
    print(llama.generate_response("What is Kubernetes?"))
