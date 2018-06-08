import requests

payload = { 
    "mount" : "/queries",
    "appInfo" : "index.js"
}

r = requests.get('http://10.0.250.2:30852/_db/voltron/_api/foxx', auth=('root', 'voltron'))
print(r.text)

z = requests.post('http://10.0.250.2:30852/_db/voltron/_api/foxx', data={'mount':'%2Fqueries'}, auth=('root', 'voltron'))  
print(z)
print(z.text)

#r = self._requests.post('http://mydev:8529/_db/mydb/_api/upload', data=data) # succeeds

#r = self._requests.put(
#    r'http://mydev:8529/_db/mydb/_admin/aardvark/foxxes/zip?mount=%2Fmy-mount-point',
#    json=data
#)




#POST /_api/foxx

#application/javascript: a standalone JavaScript file

