package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	msgFile := os.Getenv("HOME") + "/.gw_msg.txt"
	respFile := os.Getenv("HOME") + "/.gw_resp.txt"
	
	os.WriteFile(msgFile, []byte(""), 0644)
	os.WriteFile(respFile, []byte(""), 0644)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, html())
	})

	http.HandleFunc("/api/message", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Content string `json:"content"`
			ToQwen  bool   `json:"to_qwen"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		if req.ToQwen {
			// Write message
			os.WriteFile(msgFile, []byte(req.Content), 0644)
			os.WriteFile(respFile, []byte(""), 0644)
			
			// Poll for response
			for i := 0; i < 20; i++ {
				time.Sleep(500 * time.Millisecond)
				data, _ := os.ReadFile(respFile)
				if len(data) > 0 {
					os.WriteFile(respFile, []byte(""), 0644)
					json.NewEncoder(w).Encode(map[string]string{"response": strings.TrimSpace(string(data))})
					return
				}
			}
			json.NewEncoder(w).Encode(map[string]string{"response": "⏳ Type: echo 'response' > ~/.gw_resp.txt"})
			return
		}
		
		// Command mode
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, "bash", "-c", req.Content).CombinedOutput()
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"response": fmt.Sprintf("Error: %v\n%s", err, string(out))})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"response": string(out)})
	})

	fmt.Println("\n🚀 Gateway: http://localhost:9090")
	fmt.Println("💬 Messages: cat ~/.gw_msg.txt")
	fmt.Println("✍️  Respond: echo 'text' > ~/.gw_resp.txt\n")
	http.ListenAndServe(":9090", nil)
}

func html() string {
	return `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>AI Gateway</title><script src="https://cdn.tailwindcss.com"></script>
<style>.m{max-width:80%;margin:0.5rem 0;padding:0.75rem 1rem;border-radius:1rem}.mu{background:#3b82f6;color:#fff;margin-left:auto}.ma{background:#f3f4f6;margin-right:auto}.c{height:calc(100vh - 180px);overflow-y:auto}</style></head><body class="bg-gray-100">
<div class="max-w-3xl mx-auto h-screen flex flex-col">
<div class="bg-white shadow px-6 py-4"><h1 class="text-xl font-bold">💬 AI Gateway</h1><p class="text-sm text-gray-500">Chat: messages to ~/.gw_msg.txt | Respond: echo 'text' > ~/.gw_resp.txt</p></div>
<div class="bg-white border-b px-4 py-2">
<button onclick="m=true" id="b1" class="px-3 py-1 text-sm rounded bg-purple-600 text-white mr-2">💬 Chat</button>
<button onclick="m=false" id="b2" class="px-3 py-1 text-sm rounded bg-gray-200 mr-2">⚡ Cmd</button>
<button onclick="q('ls')" class="px-2 py-1 text-xs bg-gray-100 rounded">ls</button>
<button onclick="q('df -h')" class="px-2 py-1 text-xs bg-gray-100 rounded">df</button>
</div>
<div id="c" class="c flex-1 overflow-y-auto px-6 py-4"></div>
<div class="bg-white border-t px-6 py-4">
<input id="i" class="w-full border rounded-lg px-4 py-3" placeholder="Type..." onkeypress="if(event.key==='Enter')s()">
<p class="text-xs text-gray-500 mt-1" id="h">Chat mode</p>
</div>
</div>
<script>
let m=true;
function a(r,t){const d=document.createElement('div');d.className='m m'+(r==='user'?'u':'a');d.textContent=t;document.getElementById('c').appendChild(d);document.getElementById('c').scrollTop=99999;}
async function s(){const i=document.getElementById('i'),v=i.value.trim();if(!v)return;a('user',v);i.value='';i.disabled=true;try{const r=await fetch('/api/message',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({content:v,to_qwen:m})});const d=await r.json();a('assistant',d.response);}catch(e){a('assistant','Error: '+e.message);}i.disabled=false;}
function q(c){m=false;document.getElementById('i').value=c;s();}
a('assistant','Welcome! Type a message. Chat mode: write response with echo "text" > ~/.gw_resp.txt');
</script></body></html>`
}
