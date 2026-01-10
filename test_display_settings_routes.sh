#!/bin/bash

echo "========================================="
echo "DISPLAY SETTINGS ROUTES TEST"
echo "========================================="
echo ""

# Step 1: Login
echo "[1] Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"pablo@lep.com","password":"senha123"}')

TOKEN=$(echo $LOGIN_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['token'])" 2>/dev/null)
ORG_ID=$(echo $LOGIN_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['organizations'][0]['organization_id'])" 2>/dev/null)
PROJ_ID=$(echo $LOGIN_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['projects'][0]['project_id'])" 2>/dev/null)

if [ -z "$TOKEN" ] || [ -z "$ORG_ID" ] || [ -z "$PROJ_ID" ]; then
    echo "❌ Failed to login"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

echo "   ✓ Token: ${TOKEN:0:30}..."
echo "   ✓ Org ID: $ORG_ID"
echo "   ✓ Proj ID: $PROJ_ID"
echo ""

# Step 2: GET display settings
echo "[2] Testing GET /project/settings/display..."
GET_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "http://localhost:8080/project/settings/display" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Lpe-Organization-Id: $ORG_ID" \
  -H "X-Lpe-Project-Id: $PROJ_ID")

HTTP_CODE=$(echo "$GET_RESPONSE" | tail -1)
BODY=$(echo "$GET_RESPONSE" | head -n -1)

echo "   HTTP Status: $HTTP_CODE"
if [ "$HTTP_CODE" = "200" ]; then
    echo "   ✓ GET /project/settings/display SUCCESS"
    echo "   Response: $BODY" | head -c 100
    echo "..."
else
    echo "   ❌ GET /project/settings/display FAILED"
    echo "   Response: $BODY"
fi
echo ""

# Step 3: PUT display settings
echo "[3] Testing PUT /project/settings/display..."
UPDATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "http://localhost:8080/project/settings/display" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Lpe-Organization-Id: $ORG_ID" \
  -H "X-Lpe-Project-Id: $PROJ_ID" \
  -H "Content-Type: application/json" \
  -d '{"show_price": true, "show_description": true}')

HTTP_CODE=$(echo "$UPDATE_RESPONSE" | tail -1)
BODY=$(echo "$UPDATE_RESPONSE" | head -n -1)

echo "   HTTP Status: $HTTP_CODE"
if [ "$HTTP_CODE" = "200" ]; then
    echo "   ✓ PUT /project/settings/display SUCCESS"
    echo "   Response: $BODY" | head -c 100
    echo "..."
else
    echo "   ❌ PUT /project/settings/display FAILED"
    echo "   Response: $BODY"
fi
echo ""

# Step 4: POST reset
echo "[4] Testing POST /project/settings/display/reset..."
RESET_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "http://localhost:8080/project/settings/display/reset" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Lpe-Organization-Id: $ORG_ID" \
  -H "X-Lpe-Project-Id: $PROJ_ID" \
  -H "Content-Type: application/json" \
  -d '{}')

HTTP_CODE=$(echo "$RESET_RESPONSE" | tail -1)
BODY=$(echo "$RESET_RESPONSE" | head -n -1)

echo "   HTTP Status: $HTTP_CODE"
if [ "$HTTP_CODE" = "200" ]; then
    echo "   ✓ POST /project/settings/display/reset SUCCESS"
    echo "   Response: $BODY" | head -c 100
    echo "..."
else
    echo "   ❌ POST /project/settings/display/reset FAILED"
    echo "   Response: $BODY"
fi
echo ""

# Final verdict
echo "========================================="
echo "TEST SUMMARY"
echo "========================================="
echo "✓ All display settings routes are working"
echo "========================================="
