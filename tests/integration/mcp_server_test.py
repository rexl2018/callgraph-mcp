#!/usr/bin/env python3
"""
Simple test script for callgraph-mcp server
"""

import json
import subprocess
import sys

def test_mcp_server():
    # Test request
    request = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/call",
        "params": {
            "name": "callgraph",
            "arguments": {
                "moduleArgs": ["../fixtures/simple"],
                "algo": "static",
                "nostd": True
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
        
        print("STDOUT:")
        print(stdout)
        print("\nSTDERR:")
        print(stderr)
        print(f"\nReturn code: {process.returncode}")
        
        # Try to parse response
        if stdout.strip():
            try:
                response = json.loads(stdout.strip())
                print("\nParsed response:")
                print(json.dumps(response, indent=2))
            except json.JSONDecodeError as e:
                print(f"\nFailed to parse JSON response: {e}")
        
    except subprocess.TimeoutExpired:
        print("Process timed out")
        process.kill()
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    test_mcp_server()