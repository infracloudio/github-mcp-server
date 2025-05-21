body='{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "call_tool",
  "params": {
    "tool": "list_open_prs",
    "arguments": {
      "owner": "kubernetes",
      "repo": "kubernetes"
    }
  }
}'

content_length=$(echo -n "$body" | wc -c)

printf "Content-Length: %d\r\n\r\n%s" "$content_length" "$body" | go run main.go