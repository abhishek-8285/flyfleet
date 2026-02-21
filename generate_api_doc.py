import json

output = "# Comprehensive API Documentation\n\n"

try:
    with open("all_apis.json", "r") as f:
        data = json.load(f)
        apis = data.get("apis", [])
        categories = {}
        for endpoint in apis:
            path = endpoint.get("endpoint", "N/A")
            method = endpoint.get("method", "GET")
            desc = endpoint.get("description", "No description provided.")
            req = endpoint.get("request", {})
            res = endpoint.get("response", {})
            name = endpoint.get("name", "Unnamed Endpoint")
            
            parts = [p for p in path.split("/") if p]
            category = "General"
            if "api" in parts and len(parts) > 2:
                category = parts[2].capitalize()
            elif len(parts) > 0:
                category = parts[0].capitalize()
                
            categories.setdefault(category, []).append((name, method, path, desc, req, res))
            
        for cat, items in sorted(categories.items()):
            output += f"## {cat} APIs\n\n"
            for name, method, path, desc, req, res in items:
                output += f"### {name}\n"
                output += f"**Endpoint:** `[{method}] {path}`\n\n"
                output += f"**Description:** {desc}\n\n"
                if req:
                    output += "**Request Parameters / Body:**\n```json\n" + json.dumps(req, indent=2) + "\n```\n\n"
                if res:
                    output += "**Response Structure:**\n```json\n" + json.dumps(res, indent=2) + "\n```\n\n"
                output += "---\n\n"
                
    with open("/home/abhishek/.gemini/antigravity/brain/fd7ddcc1-62a5-4477-b8ae-6d0254ea3ba4/api_documentation.md", "w") as f:
        f.write(output)
    print("Successfully generated api_documentation.md")
except Exception as e:
    print(f"Error: {e}")
