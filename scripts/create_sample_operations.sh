#!/bin/bash

# Create sample operations for testing
API_URL="http://localhost:8080/api/v1"

# You may need to login first and get a session token
# Adjust the authentication as needed

echo "Creating sample Safe Provision operation..."
curl -X POST "$API_URL/operations" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "safe_provision",
    "priority": "high",
    "payload": {
      "safe_name": "Finance-Prod-Safe",
      "safe_id": "SAF_12345",
      "description": "Production safe for finance team",
      "owner": "FinanceTeam"
    }
  }'

echo -e "\n\nCreating sample Access Grant operation..."
curl -X POST "$API_URL/operations" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "access_grant",
    "priority": "medium",
    "payload": {
      "username": "john.doe@company.com",
      "user_id": "USR_67890",
      "safe_name": "Finance-Prod-Safe",
      "safe_id": "SAF_12345",
      "role": "ViewOnly",
      "reason": "Quarterly audit access"
    }
  }'

echo -e "\n\nCreating sample User Sync operation..."
curl -X POST "$API_URL/operations" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "user_sync",
    "priority": "low",
    "payload": {
      "username": "jane.smith",
      "user_id": "USR_11111",
      "email": "jane.smith@company.com",
      "source": "Active Directory",
      "domain": "CORP",
      "groups": ["Finance", "Audit"]
    }
  }'

echo -e "\n\nCreating sample Safe Delete operation..."
curl -X POST "$API_URL/operations" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "safe_delete",
    "priority": "normal",
    "payload": {
      "safe_name": "Legacy-Test-Safe",
      "safe_id": "SAF_99999",
      "reason": "Decommissioned application"
    }
  }'

echo -e "\n\nDone creating sample operations!"