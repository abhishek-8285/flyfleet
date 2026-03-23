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
	"sync"
	"time"
)

type GatewayServer struct {
	sessions     map[string]*Session
	mu           sync.RWMutex
	homeDir      string
	messageFile  string
	responseFile string
}

type Session struct {
	ID       string
	Messages []Message
	Updated  time.Time
	mu       sync.RWMutex
}

type Message struct {
	Role    string
	Content string
	Time    time.Time
}

func NewGatewayServer() *GatewayServer {
	home := os.Getenv("HOME")
	return &GatewayServer{
		sessions:     make(map[string]*Session),
		homeDir:      home,
		messageFile:  home + "/.gw_msg.txt",
		responseFile: home + "/.gw_resp.txt",
	}
}

func (s *GatewayServer) getSession(id string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[id]
}

func (s *GatewayServer) createSession() *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := &Session{ID: fmt.Sprintf("s%d", time.Now().UnixNano()), Updated: time.Now()}
	s.sessions[sess.ID] = sess
	return sess
}

func (s *GatewayServer) processMessage(sid, content string, toQwen bool) string {
	sess := s.getSession(sid)
	if sess == nil {
		return "Session not found"
	}
	
	sess.mu.Lock()
	sess.Messages = append(sess.Messages, Message{Role: "user", Content: content, Time: time.Now()})
	sess.mu.Unlock()

	if toQwen {
		// Write to file
		os.WriteFile(s.messageFile, []byte(content), 0644)
		os.WriteFile(s.responseFile, []byte(""), 0644)
		
		// Poll for response (10 sec max)
		for i := 0; i < 20; i++ {
			time.Sleep(500 * time.Millisecond)
			data, _ := os.ReadFile(s.responseFile)
			if len(data) > 0 {
				os.WriteFile(s.responseFile, []byte(""), 0644)
				resp := strings.TrimSpace(string(data))
				sess.mu.Lock()
				sess.Messages = append(sess.Messages, Message{Role: "assistant", Content: resp, Time: time.Now()})
				sess.mu.Unlock()
				return resp
			}
		}
		return "⏳ Type response in terminal: echo 'your response' > ~/.gw_resp.txt"
	}
	
	// Command mode
	sess.mu.Lock()
	defer sess.mu.Unlock()
	return s.runCommand(content)
}

func (s *GatewayServer) runCommand(cmd string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Error: %v\n%s", err, string(out))
	}
	return string(out)
}

func main() {
	srv := NewGatewayServer()
	
	// Init files
	os.WriteFile(srv.messageFile, []byte(""), 0644)
	os.WriteFile(srv.responseFile, []byte(""), 0644)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, html())
	})

	http.HandleFunc("/api/session", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(srv.createSession())
	})

	http.HandleFunc("/api/message", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			SessionID string `json:"session_id"`
			Content   string `json:"content"`
			ToQwen    bool   `json:"to_qwen"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		resp := srv.processMessage(req.SessionID, req.Content, req.ToQwen)
		json.NewEncoder(w).Encode(map[string]string{"response": resp})
	})

	fmt.Println("\n🚀 Gateway: http://localhost:9090")
	fmt.Println("💬 Messages: cat ~/.gw_msg.txt")
	fmt.Println("✍️  Respond: echo 'text' > ~/.gw_resp.txt\n")
	http.ListenAndServe(":9090", nil)
}

func html() string {
	return `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>AI Gateway</title><script src="https://cdn.tailwindcss.com"></script>
<style>.m{max-width:80%;margin:0.5rem 0;padding:0.75rem 1rem;border-radius:1rem}
.mu{background:#3b82f6;color:#fff;margin-left:auto}.ma{background:#f3f4f6;margin-right:auto}
.c{height:calc(100vh - 180px);overflow-y:auto}</style></head><body class="bg-gray-100">
<div class="max-w-3xl mx-auto h-screen flex flex-col">
<div class="bg-white shadow px-6 py-4"><h1 class="text-xl font-bold">💬 AI Gateway</h1>
<p class="text-sm text-gray-500">Chat mode: respond via terminal</p></div>
<div class="bg-white border-b px-4 py-2">
<button onclick="setM(true)" id="b1" class="px-3 py-1 text-sm rounded bg-purple-600 text-white mr-2">💬 Chat</button>
<button onclick="setM(false)" id="b2" class="px-3 py-1 text-sm rounded bg-gray-200 mr-2">⚡ Cmd</button>
<button onclick="q('ls')" class="px-2 py-1 text-xs bg-gray-100 rounded">ls</button>
<button onclick="q('df -h')" class="px-2 py-1 text-xs bg-gray-100 rounded">df</button>
</div>
<div id="c" class="c flex-1 overflow-y-auto px-6 py-4"></div>
<div class="bg-white border-t px-6 py-4">
<input id="i" class="w-full border rounded-lg px-4 py-3" placeholder="Type..." onkeypress="if(event.key==='Enter')s()">
<p class="text-xs text-gray-500 mt-1" id="h">Chat: messages to ~/.gw_msg.txt, respond to ~/.gw_resp.txt</p>
</div>
</div>
<script>
let sid=null,m=true;
async function init(){sid=(await(fetch('/api/session',{method:'POST'}).then(r=>r.json()))).ID;}
function setM(v){m=v;document.getElementById('b1').className='px-3 py-1 text-sm rounded mr-2 '+(v?'bg-purple-600 text-white':'bg-gray-200');document.getElementById('b2').className='px-3 py-1 text-sm rounded mr-2 '+(v?'bg-gray-200':'bg-purple-600 text-white');}
function a(r,t){const d=document.createElement('div');d.className='m m'+(r==='user'?'u':'a');d.textContent=t;document.getElementById('c').appendChild(d);document.getElementById('c').scrollTop=99999;}
async function s(){const i=document.getElementById('i'),v=i.value.trim();if(!v||!sid)return;a('user',v);i.value='';i.disabled=true;try{const r=await fetch('/api/message',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({session_id:sid,content:v,to_qwen:m})});const d=await r.json();a('assistant',d.response);}catch(e){a('assistant','Error: '+e.message);}i.disabled=false;}
function q(c){setM(false);document.getElementById('i').value=c;s();}
init();setM(true);a('assistant','Welcome! Type a message. In Chat mode, it goes to ~/.gw_msg.txt - respond with: echo "text" > ~/.gw_resp.txt');
</script></body></html>`
}
