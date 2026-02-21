import glob
import os

proto_files = glob.glob("proto/**/*.proto", recursive=True)

output = "\n\n## 2. gRPC APIs (Protobuf Definitions)\n\n"
output += "The following are the gRPC service definitions for the new microservice architecture:\n\n"

for fpath in sorted(proto_files):
    output += f"### {os.path.basename(fpath)}\n\n"
    output += f"**File:** `{fpath}`\n\n"
    try:
        with open(fpath, "r") as f:
            output += "```protobuf\n"
            output += f.read()
            output += "\n```\n\n"
            output += "---\n\n"
    except Exception as e:
        output += f"Error reading {fpath}: {e}\n\n"

with open("/home/abhishek/.gemini/antigravity/brain/fd7ddcc1-62a5-4477-b8ae-6d0254ea3ba4/api_documentation.md", "a") as f:
    f.write(output)

print("Appended gRPC definition to api_documentation.md")
