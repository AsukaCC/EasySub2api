#!/usr/bin/env python3
"""Validate a built image with disposable Docker PostgreSQL/Redis; no existing data.

Usage: RELEASE_TEST_IMAGE=easysub2api:release0221-test python3 scripts/test-release0221-smoke.py
"""
import os
import subprocess as sp, json, urllib.request as req, urllib.error, time, secrets, uuid, hashlib, base64, urllib.parse
opener=req.build_opener(req.ProxyHandler({}))
prefix='easy0221-'+uuid.uuid4().hex[:8]
names=[]
def docker(*args):
 return sp.check_output(['docker',*args],text=True,stderr=sp.STDOUT).strip()
def start(name,image,*args,opts=()):
 name=prefix+'-'+name; names.append(name)
 docker('run','-d','--name',name,'--network',prefix,*opts,image,*args);return name
password=secrets.token_urlsafe(24)
base='';token=''
def call(path,body=None,method=None,auth=None,expected=200,raw=False,form=False):
 headers={'Content-Type':'application/x-www-form-urlencoded' if form else 'application/json', 'Accept':'application/json, text/event-stream'}
 if auth is None:auth=token
 if auth:headers['Authorization']='Bearer '+auth
 request=req.Request(base+path,data=None if body is None else (urllib.parse.urlencode(body).encode() if form else json.dumps(body).encode()),headers=headers,method=method)
 try:r=opener.open(request,timeout=15)
 except urllib.error.HTTPError as e:r=e
 data=r.read();assert r.status==expected,(path,r.status,data[:450])
 if raw:return data
 return json.loads(data)
def ok(label):print('PASS '+label,flush=True)
try:
 docker('network','create',prefix)
 pg=start('pg','postgres:16-alpine',opts=['--tmpfs','/var/lib/postgresql/data:rw,size=512m','-e','POSTGRES_PASSWORD='+password,'-e','POSTGRES_DB=easysub2api'])
 redis=start('redis','redis:8.4-alpine')
 for _ in range(60):
  if sp.run(['docker','exec',pg,'pg_isready','-U','postgres','-d','easysub2api'],stdout=sp.DEVNULL,stderr=sp.DEVNULL).returncode==0:
   break
  time.sleep(1)
 else:
  raise RuntimeError('Temporary PostgreSQL did not become ready')
 env={'AUTO_SETUP':'true','DATABASE_HOST':pg,'DATABASE_USER':'postgres','DATABASE_PASSWORD':password,'DATABASE_DBNAME':'easysub2api','DATABASE_SSLMODE':'disable','REDIS_HOST':redis,'ADMIN_EMAIL':'release-smoke@example.com','ADMIN_PASSWORD':password,'JWT_SECRET':secrets.token_hex(32),'SERVER_MODE':'release','GOMEMLIMIT':'512MiB'}
 for enabled in [False,True]:
  if enabled:env['MCP_PUBLIC_URL']='http://localhost/api/v1/mcp'
  opts=['-p','127.0.0.1::8080']
  for k,v in env.items():opts+=['-e',k+'='+v]
  app=start('app'+str(enabled).lower(),os.environ.get('RELEASE_TEST_IMAGE','easysub2api:release0221-test'),opts=opts)
  base='http://'+docker('port',app,'8080/tcp')
  for i in range(90):
   try:
    if call('/health',auth='',raw=True):break
   except Exception:pass
   time.sleep(1)
  else:raise RuntimeError(docker('logs','--tail','60',app))
  token=call('/api/v1/auth/login',{'email':env['ADMIN_EMAIL'],'password':password},auth='')['data']['access_token']
  ok('Docker startup, all migrations, embedded web, administrator login; MCP='+str(enabled))
  assert b'<html' in call('/',auth='',raw=True).lower()
  if not enabled:
   settings=call('/api/v1/settings/public',auth='')['data'];assert not settings['mcp_enabled']
   call('/api/v1/mcp',{'jsonrpc':'2.0','id':1,'method':'initialize'},expected=503)
   compliance=call('/api/v1/admin/compliance')['data']
   call('/api/v1/admin/compliance/accept',{'language':'en','phrase':compliance['ack_phrase_en']})
   p='/api/v1/admin/settings/registration-email-blacklist'
   saved=call(p,{'emails':['smoke.test+alias@gmail.com','smoketest@gmail.com']},method='PUT')['data']
   assert saved['emails']==['smoke.test+alias@gmail.com','smoketest@gmail.com']
   assert call(p)['data']['emails']==saved['emails'];ok('Email blacklist persistence')
   price={'platform':'openai','models':['gpt-5.6-sol'],'billing_mode':'token','input_price':0.000001,'output_price':0.000002,'cache_write_price':0.000003,'cache_write_12h_price':0}
   group=call('/api/v1/admin/groups',{'name':'release0221-smoke','platform':'openai','rate_multiplier':1,'model_pricing':[price]})['data']
   uuid.UUID(group['id']);assert group['model_pricing'][0]['cache_write_12h_price']==0
   assert call('/api/v1/admin/groups/'+group['id'])['data']['model_pricing'][0]['cache_write_12h_price']==0;ok('UUID group CRUD and explicit free 12h price preservation')
   p='/api/v1/admin/settings/admin-api-keys';body={'name':'smoke','enabled':True,'policy':{'full_access':True}}
   call(p,body,expected=400)
   body['confirm_full_access']=True
   key=call(p,body)['data'];assert key['item']['policy']['full_access']
   call(p+'/'+key['item']['id'],method='DELETE');ok('Full-access key requires confirmation and supports revocation')
  else:
   config=call('/api/v1/user/mcp/config')['data'];assert config['resource']==env['MCP_PUBLIC_URL']
   client=call('/api/v1/mcp/oauth/register',{'client_name':'Release smoke','redirect_uris':['http://localhost/callback'],'token_endpoint_auth_method':'none'},auth='',expected=201)
   verifier=secrets.token_urlsafe(48);challenge=base64.urlsafe_b64encode(hashlib.sha256(verifier.encode()).digest()).decode().rstrip('=')
   authorization={'response_type':'code','client_id':client['client_id'],'redirect_uri':'http://localhost/callback','scope':'mcp','state':'test','code_challenge':challenge,'code_challenge_method':'S256','resource':env['MCP_PUBLIC_URL']}
   decision=call('/api/v1/user/mcp/authorize',{'approved':True,'authorization':authorization,'scope':'user','read_only':True})['data']
   code=urllib.parse.parse_qs(urllib.parse.urlparse(decision['redirect_url']).query)['code'][0]
   grant=call('/api/v1/mcp/oauth/token',{'grant_type':'authorization_code','client_id':client['client_id'],'redirect_uri':authorization['redirect_uri'],'code':code,'code_verifier':verifier,'resource':env['MCP_PUBLIC_URL']},auth='',form=True)
   result=call('/api/v1/mcp',{'jsonrpc':'2.0','id':1,'method':'tools/list'},auth=grant['access_token']);assert 'result' in result
   grants=call('/api/v1/user/mcp/grants')['data'];assert grants
   call('/api/v1/user/mcp/grants/'+grants[0]['id'],method='DELETE')
   call('/api/v1/mcp',{'jsonrpc':'2.0','id':2,'method':'tools/list'},auth=grant['access_token'],expected=401);ok('Enabled MCP OAuth consent, token exchange, tools, and live revocation')
  docker('rm','-f',app)
 print('All Docker application smoke checks passed.',flush=True)
finally:
 for name in reversed(names):sp.run(['docker','rm','-f',name],stdout=sp.DEVNULL,stderr=sp.DEVNULL)
 sp.run(['docker','network','rm',prefix],stdout=sp.DEVNULL,stderr=sp.DEVNULL)
