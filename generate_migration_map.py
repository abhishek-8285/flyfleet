import json
import os

def generate_plan():
    try:
        with open('all_apis.json', 'r') as f:
            data = json.load(f)
    except FileNotFoundError:
        print("Error: all_apis.json not found")
        return

    apis = data.get('apis', [])
    mapping = []
    conflicts = []
    
    # Map to track generic paths
    path_map = {} # (method, new_path) -> old_path

    for api in apis:
        old_path = api.get('endpoint')
        method = api.get('method')
        name = api.get('name')
        
        # Logic to determine new path
        new_path = old_path
        
        # Standardize prefix
        if old_path.startswith('/auth/'):
            new_path = '/api/v1' + old_path
        elif old_path.startswith('/admin/'):
            # Try to flatten admin if it makes sense, or keep it?
            # Strategy: /admin/vehicles -> /api/v1/vehicles (Resource centric)
            # Check if this collides with existing /api/ paths later.
            # specific admin overrides might stay in admin
            if old_path.startswith('/admin/users') or old_path.startswith('/admin/organizations'):
                 new_path = '/api/v1' + old_path # Keep /admin for these likely
            else:
                 new_path = '/api/v1' + old_path.replace('/admin', '', 1)
        elif old_path.startswith('/api/'):
            if not old_path.startswith('/api/v1/'):
                new_path = '/api/v1' + old_path.replace('/api', '', 1)
        elif old_path.startswith('/routes/'):
            new_path = '/api/v1/routes' + old_path.replace('/routes', '', 1)
        
        # Normalize double slashes
        new_path = new_path.replace('//', '/')
        
        # Conflict Detection
        key = (method, new_path)
        if key in path_map:
            conflicts.append({
                'path': new_path,
                'method': method,
                'conflict_with': path_map[key]['old_path'],
                'current_old': old_path
            })
        else:
            path_map[key] = {'old_path': old_path, 'name': name}
            
        mapping.append({
            'old_path': old_path,
            'new_path': new_path,
            'method': method,
            'resource': name,
            'type': 'move' if old_path != new_path else 'unchanged'
        })

    # Generate Markdown Output
    print("## 1. Migration Mapping Table")
    print("")
    print("| Original Path | New Path | Method | Resource | Type |")
    print("| :--- | :--- | :--- | :--- | :--- |")
    for row in mapping:
        print(f"| `{row['old_path']}` | `{row['new_path']}` | {row['method']} | {row['resource']} | {row['type']} |")
    
    print("")
    print("## 2. Conflicts")
    print("")
    if conflicts:
        print("| New Path | Method | Source 1 | Source 2 |")
        print("| :--- | :--- | :--- | :--- |")
        for c in conflicts:
            print(f"| `{c['path']}` | {c['method']} | `{c['conflict_with']}` | `{c['current_old']}` |")
    else:
        print("*No direct path conflicts detected based on this strategies.*")

    print("")
    print("## 3. URL Hierarchy (Preview)")
    print("```text")
    # Simple tree viz
    paths = sorted(list(set([m['new_path'] for m in mapping])))
    # Limit to top 50 for brevity if needed
    for p in paths:
        print(p)
    print("```")

if __name__ == "__main__":
    generate_plan()
