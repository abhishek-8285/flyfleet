package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		io.WriteString(w, html())
	})

	http.HandleFunc("/api/msg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		var req struct {
			M string `json:"m"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		
		reply := "Got: " + req.M
		if req.M == "hi" || req.M == "hello" {
			reply = "Hello! How can I help you today?"
		} else if req.M == "help" {
			reply = "Try: ls, df -h, ps aux, whoami, pwd, date"
		} else if len(req.M) > 1 && req.M[:2] == "ls" {
			out, _ := exec.Command("bash", "-c", req.M).CombinedOutput()
			reply = string(out)
		} else if len(req.M) > 1 && req.M[:2] == "df" {
			out, _ := exec.Command("bash", "-c", req.M).CombinedOutput()
			reply = string(out)
		} else if req.M == "ps" {
			out, _ := exec.Command("bash", "-c", "ps aux | head -15").CombinedOutput()
			reply = string(out)
		} else if req.M == "whoami" {
			out, _ := exec.Command("whoami").CombinedOutput()
			reply = string(out)
		} else if req.M == "pwd" {
			out, _ := exec.Command("pwd").CombinedOutput()
			reply = string(out)
		} else if req.M == "date" {
			out, _ := exec.Command("date").CombinedOutput()
			reply = string(out)
		}
		
		json.NewEncoder(w).Encode(map[string]string{"r": reply})
	})

	fmt.Println("\n🚀 AI Gateway: http://localhost:9090\n")
	http.ListenAndServe(":9090", nil)
}

func html() string {
	return `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>AI Gateway</title><script src="https://cdn.tailwindcss.com"></script></head>
<body class="bg-gray-100"><div class="max-w-2xl mx-auto h-screen flex flex-col">
<div class="bg-white shadow px-6 py-4"><h1 class="text-xl font-bold">💬 AI Gateway</h1><p class="text-sm text-gray-500">Type a message or command</p></div>
<div id="c" class="flex-1 overflow-y-auto px-6 py-4"></div>
<div class="bg-white border-t px-6 py-4">
<input id="i" class="w-full border rounded-lg px-4 py-3" placeholder="Type: hello, ls, df -h, help..." onkeypress="if(event.key==='Enter')s()">
</div></div><script>
function a(t,c){const d=document.createElement('div');d.className='max-w-[80%] m-2 p-3 rounded-xl '+(t==='u'?'bg-blue-500 text-white ml-auto':'bg-gray-200 mr-auto');d.textContent=c;document.getElementById('c').appendChild(d);document.getElementById('c').scrollTop=99999;}
async function s(){const i=document.getElementById('i'),v=i.value.trim();if(!v)return;a('u',v);i.value='';i.disabled=true;try{const r=await fetch('/api/msg',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({m:v})});const d=await r.json();a('a',d.r||'no response');}catch(e){a('a','Error: '+e.message);}i.disabled=false;}
a('a','Welcome! Try: hello, help, ls -la, df -h');
</script></body></html>`
}
