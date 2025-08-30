#!/bin/bash

# Test script for improved search functionality

echo "🔍 Testing ConnectSphere Search Functionality"
echo "=============================================="

# Base URL
BASE_URL="http://localhost:8080/api/v1"

# Get token (using an existing user)
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYzc1NTMxMGMtNDA4ZC00MzM3LTk0ODgtNmY2ZjU2ZGVkYTI5IiwiZW1haWwiOiJqb2huQGV4YW1wbGUuY29tIiwiZXhwIjoxNzU2NjEwMTk1LCJpYXQiOjE3NTY1MjM3OTV9.q80qrNGy2am93jLB3FrB2__2VfuyodPfv0aOvTQzlUM"

echo ""
echo "Test 1: Search for 'john' (should find john_doe)"
curl -s -X GET "${BASE_URL}/users/search?q=john&limit=10" \
  -H "Authorization: Bearer ${TOKEN}" | jq '.[].username'

echo ""
echo "Test 2: Search for 'doe' (should find john_doe by display name)"
curl -s -X GET "${BASE_URL}/users/search?q=doe&limit=10" \
  -H "Authorization: Bearer ${TOKEN}" | jq '.[].display_name'

echo ""
echo "Test 3: Search for 'jane' (should find jane_smith)"
curl -s -X GET "${BASE_URL}/users/search?q=jane&limit=10" \
  -H "Authorization: Bearer ${TOKEN}" | jq '.[].username'

echo ""
echo "Test 4: Search for 'smith' (should find jane_smith by display name)"
curl -s -X GET "${BASE_URL}/users/search?q=smith&limit=10" \
  -H "Authorization: Bearer ${TOKEN}" | jq '.[].display_name'

echo ""
echo "Test 5: Partial search 'j' (should find both john and jane, ranked appropriately)"
curl -s -X GET "${BASE_URL}/users/search?q=j&limit=10" \
  -H "Authorization: Bearer ${TOKEN}" | jq '.[] | {username: .username, display_name: .display_name}'

echo ""
echo "✅ Search functionality testing complete!"
echo ""
echo "🎯 Key Improvements:"
echo "  - Partial matching on both username and display name"
echo "  - Smart ranking: exact matches first, then prefix matches, then partial matches"
echo "  - Case-insensitive search"
echo "  - Shorter names ranked higher for better relevance"
echo "  - Debounced search in Flutter (500ms delay) to reduce API calls"
echo "  - Minimum 2 characters required for search to start"
