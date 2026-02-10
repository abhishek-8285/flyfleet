#!/usr/bin/env python3

import os
import re
import sys
from collections import defaultdict

def analyze_go_file(filepath):
    """Analyze a Go file for hardcoded values and issues"""
    issues = []
    
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
            lines = content.split('\n')
    except Exception as e:
        return [f"Error reading file: {str(e)}"]
    
    # Check for hardcoded database credentials
    db_patterns = [
        r'postgres://[^:]+:[^@]+@',
        r'DATABASE_URL\s*=\s*["\'][^"\']+["\']',
        r'password\s*=\s*["\'][^"\']+["\']',
        r'username\s*=\s*["\'][^"\']+["\']',
    ]
    
    for i, line in enumerate(lines, 1):
        # Hardcoded secrets/passwords
        if re.search(r'(password|secret|token|key)\s*[:=]\s*["\'][^"\']{5,}["\']', line, re.IGNORECASE):
            issues.append(f"Line {i}: Hardcoded secret/password found")
        
        # Hardcoded URLs/hosts
        if re.search(r'(localhost|127\.0\.0\.1|0\.0\.0\.0):\d+', line) and not re.search(r'os\.Getenv|config\.|ENV', line):
            issues.append(f"Line {i}: Hardcoded host/port found")
            
        # Hardcoded file paths
        if re.search(r'/tmp/|/var/|/home/|/Users/', line) and not re.search(r'os\.Getenv|config\.|ENV', line):
            issues.append(f"Line {i}: Hardcoded file path found")
            
        # Hardcoded numeric constants (magic numbers)
        # Skip if it's in a const block or obvious constants
        if re.search(r'\b\d{3,}\b', line) and not re.search(r'const\s|version|port|HTTP|Status|time\.', line):
            # Check if it's assigned to a variable (likely magic number)
            if re.search(r'[a-zA-Z_][a-zA-Z0-9_]*\s*[:=]\s*\d{3,}', line):
                issues.append(f"Line {i}: Magic number found")
                
        # Hardcoded strings that should be configurable
        if re.search(r'["\'][a-zA-Z]{10,}["\']\s*$', line.strip()) and not re.search(r'const\s|var\s|ENV|config', line):
            issues.append(f"Line {i}: Hardcoded string constant found")
            
        # Check database patterns
        for pattern in db_patterns:
            if re.search(pattern, line, re.IGNORECASE):
                issues.append(f"Line {i}: Hardcoded database credential found")
                
        # Hardcoded API keys
        if re.search(r'[A-Za-z0-9_-]{30,}', line) and re.search(r'(key|api|token)', line, re.IGNORECASE):
            issues.append(f"Line {i}: Potential hardcoded API key found")
            
        # Hardcoded email addresses
        if re.search(r'[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}', line) and not re.search(r'ENV|config|test', line):
            issues.append(f"Line {i}: Hardcoded email address found")
            
    return issues

def main():
    go_files_path = '/home/abhishek/Desktop/flyfleet/go_files_list.txt'
    
    if not os.path.exists(go_files_path):
        print("Go files list not found. Run find command first.")
        return
    
    with open(go_files_path, 'r') as f:
        files = [line.strip() for line in f.readlines()]
    
    print(f"Analyzing {len(files)} Go files...\n")
    
    total_issues = 0
    files_with_issues = 0
    
    # Group issues by service
    service_issues = defaultdict(list)
    
    for filepath in files:
        if not filepath:
            continue
            
        # Extract service name
        service_match = re.search(r'services/([^/]+)/', filepath)
        if service_match:
            service_name = service_match.group(1)
        else:
            service_name = "common" if "packages/common" in filepath else "other"
            
        issues = analyze_go_file(filepath)
        if issues:
            files_with_issues += 1
            total_issues += len(issues)
            service_issues[service_name].append((filepath, issues))
    
    # Print summary
    print("=" * 80)
    print("GO CODE ANALYSIS SUMMARY")
    print("=" * 80)
    print(f"Total Go files analyzed: {len(files)}")
    print(f"Files with issues: {files_with_issues}")
    print(f"Total issues found: {total_issues}")
    print()
    
    # Print detailed results by service
    for service, file_issues in sorted(service_issues.items()):
        print(f"\n--- {service.upper()} SERVICE ---")
        print("-" * 40)
        
        for filepath, issues in file_issues:
            print(f"\n📁 {filepath}")
            for issue in issues:
                print(f"  ❌ {issue}")
    
    print("\n" + "=" * 80)
    print("ANALYSIS COMPLETE")

if __name__ == "__main__":
    main()
