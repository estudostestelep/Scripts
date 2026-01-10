#!/bin/bash

# Login
curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"pablo@lep.com","password":"senha123"}' > /tmp/login.json

# Extract credentials
TOKEN=$(cat /tmp/login.json | python3 -c "import sys, json; print(json.load(sys.stdin)['token'])")
ORG_ID=$(cat /tmp/login.json | python3 -c "import sys, json; print(json.load(sys.stdin)['organizations'][0]['organization_id'])")
PROJ_ID=$(cat /tmp/login.json | python3 -c "import sys, json; print(json.load(sys.stdin)['projects'][0]['project_id'])")

echo "Credentials obtained"
echo ""

# GET theme
echo "=== GET /project/settings/theme ==="
curl -s -X GET "http://localhost:8080/project/settings/theme" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Lpe-Organization-Id: $ORG_ID" \
  -H "X-Lpe-Project-Id: $PROJ_ID" | tee /tmp/get_theme.txt
echo ""
echo ""

# Reset theme
echo "=== POST /project/settings/theme/reset ==="
curl -s -X POST "http://localhost:8080/project/settings/theme/reset" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Lpe-Organization-Id: $ORG_ID" \
  -H "X-Lpe-Project-Id: $PROJ_ID" \
  -H "Content-Type: application/json" \
  -d '{}' | tee /tmp/reset_theme.txt
echo ""
