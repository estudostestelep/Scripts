#!/bin/bash

echo "========================================="
echo "THEME RESET FUNCTIONALITY TEST"
echo "========================================="
echo ""

# Step 1: Login
echo "[1] Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"pablo@lep.com","password":"senha123"}')

TOKEN=$(echo $LOGIN_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['token'])")
ORG_ID=$(echo $LOGIN_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['organizations'][0]['organization_id'])")
PROJ_ID=$(echo $LOGIN_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['projects'][0]['project_id'])")

echo "   ✓ Token: ${TOKEN:0:30}..."
echo "   ✓ Org ID: $ORG_ID"
echo "   ✓ Proj ID: $PROJ_ID"
echo ""

# Step 2: Get current theme
echo "[2] Getting current theme..."
GET_RESPONSE=$(curl -s -X GET "http://localhost:8080/project/settings/theme" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Lpe-Organization-Id: $ORG_ID" \
  -H "X-Lpe-Project-Id: $PROJ_ID")

echo "$GET_RESPONSE" | python3 -m json.tool
echo ""

# Step 3: Reset theme
echo "[3] Resetting theme to defaults..."
RESET_RESPONSE=$(curl -s -X POST "http://localhost:8080/project/settings/theme/reset" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Lpe-Organization-Id: $ORG_ID" \
  -H "X-Lpe-Project-Id: $PROJ_ID" \
  -H "Content-Type: application/json" \
  -d '{}')

echo "$RESET_RESPONSE" | python3 -m json.tool
echo ""

# Step 4: Verify reset values
echo "[4] Verifying reset values..."
PRIMARY_LIGHT=$(echo "$RESET_RESPONSE" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('data', {}).get('primary_color_light', 'NOT_FOUND'))")
PRIMARY_DARK=$(echo "$RESET_RESPONSE" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('data', {}).get('primary_color_dark', 'NOT_FOUND'))")
BG_LIGHT=$(echo "$RESET_RESPONSE" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('data', {}).get('background_color_light', 'NOT_FOUND'))")
BG_DARK=$(echo "$RESET_RESPONSE" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('data', {}).get('background_color_dark', 'NOT_FOUND'))")

echo "   Primary Light: $PRIMARY_LIGHT (expected: #1E293B)"
echo "   Primary Dark: $PRIMARY_DARK (expected: #F8FAFC)"
echo "   Background Light: $BG_LIGHT (expected: #FFFFFF)"
echo "   Background Dark: $BG_DARK (expected: #0F172A)"
echo ""

# Final verdict
if [[ "$PRIMARY_LIGHT" == "#1E293B" ]] && [[ "$PRIMARY_DARK" == "#F8FAFC" ]] && [[ "$BG_LIGHT" == "#FFFFFF" ]] && [[ "$BG_DARK" == "#0F172A" ]]; then
    echo "========================================="
    echo "✅ THEME RESET TEST PASSED"
    echo "========================================="
    exit 0
else
    echo "========================================="
    echo "❌ THEME RESET TEST FAILED"
    echo "========================================="
    exit 1
fi
