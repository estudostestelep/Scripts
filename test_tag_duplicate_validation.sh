#!/bin/bash

echo "========================================="
echo "TAG DUPLICATE VALIDATION TEST"
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

# Step 2: Create first tag
echo "[2] Creating first tag (should succeed)..."
TAG_1_RESPONSE=$(curl -s -X POST "http://localhost:8080/tag" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Lpe-Organization-Id: $ORG_ID" \
  -H "X-Lpe-Project-Id: $PROJ_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "premium",
    "entity_type": "product",
    "color": "#FF5733"
  }')

echo "$TAG_1_RESPONSE" | python3 -m json.tool
TAG_1_ID=$(echo "$TAG_1_RESPONSE" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('data', {}).get('id', 'ERROR'))" 2>/dev/null)
echo ""

if [ "$TAG_1_ID" = "ERROR" ] || [ -z "$TAG_1_ID" ]; then
    echo "❌ Failed to create first tag"
    exit 1
fi
echo "   ✓ Tag 1 ID: $TAG_1_ID"
echo ""

# Step 3: Try to create duplicate tag (should fail)
echo "[3] Creating duplicate tag with same name+type (should fail)..."
TAG_2_RESPONSE=$(curl -s -X POST "http://localhost:8080/tag" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Lpe-Organization-Id: $ORG_ID" \
  -H "X-Lpe-Project-Id: $PROJ_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "premium",
    "entity_type": "product",
    "color": "#00FF00"
  }')

echo "$TAG_2_RESPONSE" | python3 -m json.tool
ERROR_MESSAGE=$(echo "$TAG_2_RESPONSE" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('error', 'NO_ERROR'))" 2>/dev/null)
echo ""

# Step 4: Create tag with same name but different type (should succeed)
echo "[4] Creating tag with same name but different type (should succeed)..."
TAG_3_RESPONSE=$(curl -s -X POST "http://localhost:8080/tag" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Lpe-Organization-Id: $ORG_ID" \
  -H "X-Lpe-Project-Id: $PROJ_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "premium",
    "entity_type": "customer",
    "color": "#0000FF"
  }')

echo "$TAG_3_RESPONSE" | python3 -m json.tool
TAG_3_ID=$(echo "$TAG_3_RESPONSE" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('data', {}).get('id', 'ERROR'))" 2>/dev/null)
echo ""

if [ "$TAG_3_ID" = "ERROR" ] || [ -z "$TAG_3_ID" ]; then
    echo "⚠️  Could not create tag with same name but different type"
else
    echo "   ✓ Tag 3 ID: $TAG_3_ID (same name, different type - success)"
fi
echo ""

# Step 5: Try to update tag 1 to match tag 3 (should fail)
echo "[5] Updating tag 1 to have same name+type as tag 3 (should fail)..."
UPDATE_RESPONSE=$(curl -s -X PUT "http://localhost:8080/tag/$TAG_1_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Lpe-Organization-Id: $ORG_ID" \
  -H "X-Lpe-Project-Id: $PROJ_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "'$TAG_1_ID'",
    "name": "premium",
    "entity_type": "customer",
    "color": "#FF0000"
  }')

echo "$UPDATE_RESPONSE" | python3 -m json.tool
UPDATE_ERROR=$(echo "$UPDATE_RESPONSE" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('error', 'NO_ERROR'))" 2>/dev/null)
echo ""

# Final verdict
echo "========================================="
echo "TEST RESULTS:"
echo "========================================="
echo "✓ First tag created successfully: $TAG_1_ID"

if [[ "$ERROR_MESSAGE" == *"already exists"* ]]; then
    echo "✓ Duplicate tag creation blocked correctly"
else
    echo "❌ Duplicate tag creation was NOT blocked"
fi

if [ ! -z "$TAG_3_ID" ] && [ "$TAG_3_ID" != "ERROR" ]; then
    echo "✓ Tag with same name but different type created: $TAG_3_ID"
else
    echo "⚠️  Could not verify same-name different-type scenario"
fi

if [[ "$UPDATE_ERROR" == *"already exists"* ]]; then
    echo "✓ Update to duplicate name+type blocked correctly"
elif [[ "$UPDATE_ERROR" == "NO_ERROR" ]]; then
    echo "❌ Update to duplicate name+type was NOT blocked"
fi

echo "========================================="
