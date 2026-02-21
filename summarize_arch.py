import json
import os
import glob

output = "# Comprehensive Platform Discovery: APIs, Pages, and Screens\n\n"

# APIS
output += "## 1. Backend APIs (`all_apis.json`)\n\n"
try:
    with open("all_apis.json", "r") as f:
        data = json.load(f)
        apis = data.get("apis", [])
        categories = {}
        for endpoint in apis:
            path = endpoint.get("endpoint", "N/A")
            method = endpoint.get("method", "GET")
            # Group by first part of path
            parts = [p for p in path.split("/") if p]
            category = "General"
            if "api" in parts and len(parts) > 2:
                category = parts[2].capitalize()
            elif len(parts) > 0:
                category = parts[0].capitalize()
            categories.setdefault(category, []).append(f"- **[{method}]** `{path}`")
            
        for cat, items in sorted(categories.items()):
            output += f"### {cat} APIs\n"
            for item in items:
                output += f"{item}\n"
            output += "\n"
except Exception as e:
    output += f"Error loading APIs: {e}\n"

# DASHBOARD PAGES
output += "## 2. Dashboard Web Pages (`apps/dashboard-web/src/pages`)\n\n"
dash_files = glob.glob("apps/dashboard-web/src/pages/**/*.tsx", recursive=True)
for f in sorted(dash_files):
    name = f.replace("apps/dashboard-web/src/pages/", "")
    output += f"- `{name}`\n"

# LANDING PAGE ROUTES
output += "\n## 3. Landing Page Routes (`apps/landing-page/src/app`)\n\n"
landing_files = glob.glob("apps/landing-page/src/app/**/page.tsx", recursive=True)
for f in sorted(landing_files):
    name = f.replace("apps/landing-page/src/app/", "").replace("/page.tsx", "").replace("page.tsx", "(Home)")
    if name == "": name = "(Home)"
    if name != "(Home)" and not name.startswith("/"): name = "/" + name
    output += f"- `{name}`\n"

# MOBILE APP SCREENS
output += "\n## 4. Mobile App Screens (`apps/mobile/vehicle-device/android-device/app/src/main/kotlin/com/flyfleet/vehicledevice/ui`)\n\n"
mobile_files = glob.glob("apps/mobile/vehicle-device/android-device/app/src/main/kotlin/com/flyfleet/vehicledevice/ui/**/*.kt", recursive=True)
for f in sorted(mobile_files):
    if "Screen" in f or "Activity" in f or "Fragment" in f:
        name = f.replace("apps/mobile/vehicle-device/android-device/app/src/main/kotlin/com/flyfleet/vehicledevice/ui/", "")
        output += f"- `{name}`\n"

with open("/home/abhishek/.gemini/antigravity/brain/fd7ddcc1-62a5-4477-b8ae-6d0254ea3ba4/architecture_details.md", "w") as f:
    f.write(output)

print("Created architecture_details.md")
