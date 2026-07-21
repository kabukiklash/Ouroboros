import urllib.request
import json

url = "http://127.0.0.1:3000/v1/gpp/ingest"
payload = {
    "workflow": "PythonSandboxTest",
    "type": "PYTHON_AGENT",
    "retention": "EPHEMERAL",
    "events": [
        {
            "name": "TEST_EXECUTION",
            "commands": [
                { "action": "set", "target": "state", "value": "RUNNING" },
                { "action": "increase", "target": "friction", "value": 2 }
            ]
        }
    ]
}

data = json.dumps(payload).encode("utf-8")
req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/json"}, method="POST")

try:
    with urllib.request.urlopen(req) as response:
        resp_data = response.read().decode("utf-8")
        print("✅ SUCESSO: Sandbox aceitou o GPP via Python!")
        print("Resposta do GenesisCore:")
        print(json.dumps(json.loads(resp_data), indent=2))
except urllib.error.URLError as e:
    print("❌ ERRO ao conectar ou processar:", e)
    if hasattr(e, 'read'):
        print(e.read().decode("utf-8"))
