import json
import os

output = "# PostgreSQL Database Schema Reference\n\n"

try:
    with open("db_schema.json", "r") as f:
        data = json.load(f)
        tables = data.get("tables", [])
        
        output += f"Total Tables: {len(tables)}\n\n"
        
        for table in tables:
            name = table.get("name", "Unknown")
            desc = table.get("description", "No description")
            columns = table.get("columns", [])
            
            output += f"## Table: `{name}`\n"
            output += f"**Description:** {desc}\n\n"
            
            output += "| Column | Type | Nullable | Description |\n"
            output += "|---|---|---|---|\n"
            for col in columns:
                col_name = col.get("name", "N/A")
                col_type = col.get("type", "N/A")
                col_null = "Yes" if col.get("nullable", False) else "No"
                col_desc = col.get("description", "")
                
                # Highlight primary/foreign keys in description if present internally
                if col.get("primary_key", False):
                    col_desc = f"**[PK]** {col_desc}"
                if col.get("foreign_key", False):
                    fk_target = col.get("references", "")
                    col_desc = f"**[FK -> {fk_target}]** {col_desc}"
                    
                output += f"| `{col_name}` | `{col_type}` | {col_null} | {col_desc} |\n"
                
            output += "\n---\n\n"

    with open("/home/abhishek/.gemini/antigravity/brain/fd7ddcc1-62a5-4477-b8ae-6d0254ea3ba4/database_schema.md", "w") as f:
        f.write(output)
    
    print("Database schema markdown created successfully.")

except Exception as e:
    print(f"Error parsing database schema: {e}")
