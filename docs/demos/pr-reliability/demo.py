"""Exercise both CLIs against a local fixture; no live write or credential."""
import json,subprocess,sys,threading
from http.server import BaseHTTPRequestHandler,ThreadingHTTPServer
from urllib.parse import urlsplit
mode='';requests=[];reads=0
class API(BaseHTTPRequestHandler):
 def log_message(self,*args):pass
 def reply(self,code,data):
  self.send_response(code);self.send_header('Content-Type','application/json');self.end_headers();self.wfile.write(json.dumps(data).encode())
 def do_GET(self):
  global reads
  path=urlsplit(self.path).path
  if path=='/api/v1/version':return self.reply(200,{'version':'1.24.6'})
  if path=='/api/v1/repos/example/demo':return self.reply(200,{'private':False})
  requests.append((self.command,path))
  if path.endswith('/assets'):return self.reply(200,[])
  reads+=1
  if mode=='transient-read'and reads==1:return self.reply(503,{'message':'fixture gateway unavailable'})
  return self.reply(200,None if mode=='null-read' else {'number':7,'title':'fixture','state':'open'})
 def do_POST(self):
  requests.append((self.command,urlsplit(self.path).path));self.rfile.read(int(self.headers.get('Content-Length','0')))
  if mode=='up-to-date':return self.reply(500,{'message':'HeadBranch of PR 7 is up to date'})
  if mode=='git-stderr':return self.reply(500,{'message':'exit status 128: fixture lock exists'})
  return self.reply(503,{'message':'fixture gateway unavailable'})
srv=ThreadingHTTPServer(('127.0.0.1',0),API);thread=threading.Thread(target=srv.serve_forever,daemon=True);thread.start()
def rpc(p,i,method,params):
 p.stdin.write(json.dumps({'jsonrpc':'2.0','id':i,'method':method,'params':params})+'\n');p.stdin.flush()
 while True:
  line=p.stdout.readline()
  if not line:raise RuntimeError('CLI closed')
  r=json.loads(line)
  if r.get('id')==i:return r

def probe(binary,candidate):
 global mode,reads
 p=subprocess.Popen([binary,'-host',f'http://127.0.0.1:{srv.server_port}','-tools','pull_request_read,pull_request_write'],stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=subprocess.DEVNULL,text=True)
 try:
  rpc(p,1,'initialize',{'protocolVersion':'2025-03-26','capabilities':{},'clientInfo':{'name':'pr-reliability-demo','version':'1'}})
  p.stdin.write(json.dumps({'jsonrpc':'2.0','method':'notifications/initialized'})+'\n');p.stdin.flush()
  print('CANDIDATE'if candidate else'BASELINE',flush=True)
  for i,case in enumerate(['up-to-date','git-stderr','write-gateway','null-read','transient-read','valid-read'],2):
   mode=case;reads=0;requests.clear();write=case in ['up-to-date','git-stderr','write-gateway']
   args={'owner':'example','repo':'demo','pull_number':7,'method':'update_branch'if write else'get'}
   r=rpc(p,i,'tools/call',{'name':'pull_request_write'if write else'pull_request_read','arguments':args})['result']
   text=' '.join(x.get('text','')for x in r.get('content',[]));error=bool(r.get('isError'))
   print(json.dumps({'case':case,'error':error,'response':text,'fixture_requests':len(requests),'writes':sum(m=='POST'for m,_ in requests)}),flush=True)
   expected={'up-to-date':not candidate,'git-stderr':True,'write-gateway':True,'null-read':candidate,'transient-read':not candidate,'valid-read':False}[case]
   assert error==expected,(case,r)
   if write:assert len(requests)==1,(case,requests)
   if candidate and case=='null-read':assert len(requests)==1
   if case=='transient-read':assert reads==(2 if candidate else 1)
 finally:p.stdin.close();p.wait(timeout=10)
try:probe(sys.argv[1],False);probe(sys.argv[2],True)
finally:srv.shutdown();srv.server_close();thread.join(timeout=5)
