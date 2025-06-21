#!/bin/bash

# Grafana Dashboard Import Script
# This script automatically imports the MM-Rules dashboard into Grafana

set -e

# Configuration
GRAFANA_URL="http://localhost:3000"
GRAFANA_USER="admin"
GRAFANA_PASSWORD="nxz7pvq@gem2fvf6FQT"
DASHBOARD_FILE="monitoring/grafana-dashboard.json"

TMP_PAYLOAD="/tmp/grafana_dashboard_payload.json"

echo "🚀 Importing MM-Rules Dashboard to Grafana..."
echo "=============================================="

# Wait for Grafana to be ready
echo "⏳ Waiting for Grafana to be ready..."
until curl -s "$GRAFANA_URL/api/health" > /dev/null 2>&1; do
    echo "   Waiting for Grafana..."
    sleep 2
done
echo "✅ Grafana is ready!"

# Check if dashboard file exists
if [ ! -f "$DASHBOARD_FILE" ]; then
    echo "❌ Dashboard file not found: $DASHBOARD_FILE"
    exit 1
fi

echo "📋 Importing dashboard from: $DASHBOARD_FILE"

# Read the dashboard JSON and wrap it in the correct API format
DASH_JSON=$(cat "$DASHBOARD_FILE")
echo '{"dashboard":' > "$TMP_PAYLOAD"
echo "$DASH_JSON" >> "$TMP_PAYLOAD"
echo ', "overwrite": true}' >> "$TMP_PAYLOAD"

# Import the dashboard
RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -u "$GRAFANA_USER:$GRAFANA_PASSWORD" \
    -d @"$TMP_PAYLOAD" \
    "$GRAFANA_URL/api/dashboards/db")

# Clean up temp file
rm -f "$TMP_PAYLOAD"

# Check if import was successful
if echo "$RESPONSE" | grep -q '"status":"success"'; then
    DASHBOARD_URL=$(echo "$RESPONSE" | jq -r '.url' 2>/dev/null || echo "")
    echo "✅ Dashboard imported successfully!"
    echo "🌐 Dashboard URL: $GRAFANA_URL$DASHBOARD_URL"
    echo ""
    echo "🎉 Your MM-Rules monitoring dashboard is now ready!"
    echo "   - Grafana: $GRAFANA_URL"
    echo "   - Prometheus: http://localhost:9090"
    echo "   - Matchmaking API: http://localhost:8080"
else
    echo "❌ Failed to import dashboard"
    echo "Response: $RESPONSE"
    exit 1
fi

echo ""
echo "📊 Dashboard includes:"
echo "   • Match requests rate"
echo "   • Matches created rate" 
echo "   • Queue size monitoring"
echo "   • Allocation request rates"
echo "   • Processing times"
echo "   • HTTP request rates"
echo ""
echo "💡 Run 'make demo' to generate metrics data!" 