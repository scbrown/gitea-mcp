"""Local MCP before/after with a private HTTP fixture; no credential or live API."""
import json,subprocess,sys,threading
from http.server import BaseHTTPRequestHandler,ThreadingHTTPServer
from urllib.parse import urlsplit,parse_qs
calls=[]
class API(BaseHTTPRequestHandler):
 def log_message(self,*args):pass
 def do_GET(self):
  path=urlsplit(self.path);calls.append((self.command,path.path))
  code=200
  if path.path=='/api/v1/version':data={'version':'1.24.6'}
  elif path.path=='/api/v1/repos/example/demo':data={'private':False}
  elif path.path.endswith('/actions/runs'):
   if parse_qs(path.query).get('status')==['in_progress']:code,data=400,{'message':'unknown status: in_progress'}
   else:data={'total_count':1,'workflow_runs':[{'id':42,'status':'running'}]}
  elif path.path=='/api/v1/repos/example/demo/pulls':data=[{'number':7,'title':'fixture','state':'open'}]
  else:code,data=404,{'message':'unexpected fixture path'}
  self.send_response(code);self.send_header('Content-Type','application/json');self.end_headers();self.wfile.write(json.dumps(data).encode())
 def do_PATCH(self):
  calls.append((self.command,urlsplit(self.path).path));self.rfile.read(int(self.headers.get('Content-Length','0')))
  self.send_response(200);self.send_header('Content-Type','application/json');self.end_headers();self.wfile.write(b'{"number":7,"title":"updated","state":"open"}')
server=ThreadingHTTPServer(('127.0.0.1',0),API)
thread=threading.Thread(target=server.serve_forever,daemon=True);thread.start()
def request(p,i,method,params):
 p.stdin.write(json.dumps({'jsonrpc':'2.0','id':i,'method':method,'params':params})+'\n');p.stdin.flush()
 while True:
  line=p.stdout.readline()
  if not line:raise RuntimeError('CLI closed')
  r=json.loads(line)
  if r.get('id')==i:return r

def probe(binary,candidate):
 args=[binary,'-host',f'http://127.0.0.1:{server.server_port}','-tools','actions_run_read,actions_run_write,pull_request_read,pull_request_write']
 if candidate:args+=['-actions-profile','forgejo-15']
 p=subprocess.Popen(args,stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=subprocess.DEVNULL,text=True)
 try:
  request(p,1,'initialize',{'protocolVersion':'2025-03-26','capabilities':{},'clientInfo':{'name':'call-shape-demo','version':'1'}})
  p.stdin.write(json.dumps({'jsonrpc':'2.0','method':'notifications/initialized'})+'\n');p.stdin.flush()
  print('CANDIDATE'if candidate else'BASELINE',flush=True)
  base={'owner':'example','repo':'demo'}
  cases=[('actions_run_read',{}),('actions_run_read',dict(base,method='list',status='in_progress')),('actions_run_read',dict(base,method='jobs')),('pull_request_read',dict(base,method='get')),('pull_request_read',dict(base,method='list')),('pull_request_write',dict(base,method='edit',pull_number=7,title='updated')),('actions_run_write',dict(base,method='dispatch'))]
  for i,(name,params)in enumerate(cases,2):
   before=len(calls);r=request(p,i,'tools/call',{'name':name,'arguments':params})['result']
   text=' '.join(c.get('text','')for c in r.get('content',[]))
   print(json.dumps({'tool':name,'args':params,'error':bool(r.get('isError')),'response':text,'fixture_http_requests':len(calls)-before}),flush=True)
   expected=candidate and params.get('method')in ['list','edit']
   assert bool(r.get('isError')) != expected,(name,params,r)
   if not expected:assert len(calls)==before,(name,calls)
 finally:p.stdin.close();p.wait(timeout=10)
try:probe(sys.argv[1],False);probe(sys.argv[2],True)
finally:server.shutdown();server.server_close();thread.join(timeout=5)
