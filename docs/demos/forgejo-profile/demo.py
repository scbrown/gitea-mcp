import json, subprocess, sys

def call(p, request):
 p.stdin.write(json.dumps(request)+'\n');p.stdin.flush()
 while True:
  line=p.stdout.readline()
  if not line: raise RuntimeError('CLI closed before response')
  response=json.loads(line)
  if response.get('id')==request['id']: return response

def probe(binary, profile):
 args=[binary,'-host','https://example.invalid','-tools','actions_run_read,actions_run_write']
 if profile: args+=['-actions-profile',profile]
 p=subprocess.Popen(args,stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=subprocess.DEVNULL,text=True)
 try:
  response=call(p,{'jsonrpc':'2.0','id':1,'method':'initialize','params':{'protocolVersion':'2025-03-26','capabilities':{},'clientInfo':{'name':'capability-demo','version':'1'}}})
  assert 'result' in response,response
  p.stdin.write(json.dumps({'jsonrpc':'2.0','method':'notifications/initialized'})+'\n');p.stdin.flush()
  response=call(p,{'jsonrpc':'2.0','id':2,'method':'tools/list','params':{}})
  tools=response['result']['tools']
  methods={t['name']:t['inputSchema']['properties']['method']['enum'] for t in tools}
  assert set(methods)=={'actions_run_read','actions_run_write'},methods
  print(json.dumps({'profile':profile or 'original v1.7.0','methods':methods},indent=2),flush=True)
  if profile:
   assert methods['actions_run_read']==['list_runs','get_run']
   assert methods['actions_run_write']==['dispatch_workflow']
   for number,(tool,method) in enumerate([('actions_run_read','list_run_jobs'),('actions_run_read','get_job_log_preview'),('actions_run_write','rerun_run')],3):
    response=call(p,{'jsonrpc':'2.0','id':number,'method':'tools/call','params':{'name':tool,'arguments':{'method':method}}})
    assert response['result']['isError'],response
    text=' '.join(c.get('text','') for c in response['result']['content'])
    assert 'No request sent' in text,text
    print(method+': '+text,flush=True)
  else:
   assert 'list_run_jobs' in methods['actions_run_read']
   assert 'get_job_log_preview' in methods['actions_run_read']
   assert 'rerun_run' in methods['actions_run_write']
 finally:
  p.stdin.close()
  try:p.wait(timeout=10)
  except subprocess.TimeoutExpired:p.terminate();p.wait(timeout=10)

probe(sys.argv[1],None)
probe(sys.argv[2],'forgejo-15')
