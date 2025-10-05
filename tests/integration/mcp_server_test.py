#!/usr/bin/env python3
"""
Simple test script for callgraph-mcp server
"""

import json
import subprocess
import sys

def test_mcp_server():
    # Test request with package grouping
    request = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/call",
        "params": {
            "name": "callHierarchy",
            "arguments": {
                "moduleArgs": ["../fixtures/simple"],
                "algo": "static",
                "nostd": True,
                "group": "pkg"
            }
        }
    }
    
    try:
        # Start the MCP server
        process = subprocess.Popen(
            ["../../callgraph-mcp"],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        
        # Send request
        request_str = json.dumps(request) + "\n"
        stdout, stderr = process.communicate(input=request_str, timeout=30)
        
        print("=== MCP Request ===")
        print(json.dumps(request, indent=2))
        
        print("\n=== MCP Response ===")
        print("STDOUT:")
        print(stdout)
        print("\nSTDERR:")
        print(stderr)
        print(f"\nReturn code: {process.returncode}")
        
        # Try to parse response and extract Mermaid content
        if stdout.strip():
            try:
                response = json.loads(stdout.strip())
                print("\n=== Parsed JSON Response ===")
                print(json.dumps(response, indent=2))
                
                # Extract and display Mermaid content
                if "result" in response and "content" in response["result"]:
                    content = response["result"]["content"]
                    if content and len(content) > 0 and "text" in content[0]:
                        mermaid_text = content[0]["text"]
                        print("\n=== Mermaid Flowchart Output ===")
                        print(mermaid_text)
                        print("=== End of Mermaid Output ===")
                        
            except json.JSONDecodeError as e:
                print(f"\nFailed to parse JSON response: {e}")
        
    except subprocess.TimeoutExpired:
        print("Process timed out")
        process.kill()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    test_mcp_server()