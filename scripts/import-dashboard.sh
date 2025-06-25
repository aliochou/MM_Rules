#!/bin/bash

# Grafana Dashboard Import Script
# This script automatically imports the MM-Rules dashboard into Grafana

set -e

# Configuration
GRAFANA_URL="http://localhost:3000"
GRAFANA_USER="admin"
GRAFANA_PASS="nxz7pvq@gem2fvf6FQT"
DASHBOARD_PATH="monitoring/grafana-dashboard.json"
DATASOURCE_NAME="Prometheus"

TMP_PAYLOAD="/tmp/grafana_dashboard_payload.json"

echo "🚀 Importing MM-Rules Dashboard to Grafana..."
echo "=============================================="

# Wait for Grafana to be ready
echo "⏳ Waiting for Grafana to be ready..."
until curl -s -u "$GRAFANA_USER:$GRAFANA_PASS" "$GRAFANA_URL/api/health" | grep -q '"database": "ok"'; do
    echo -n "."
    sleep 1
done
echo "✅ Grafana is ready!"

# Check if dashboard file exists
if [ ! -f "$DASHBOARD_PATH" ]; then
    echo "❌ Dashboard file not found: $DASHBOARD_PATH"
    exit 1
fi

# Get Prometheus datasource UID from Grafana
echo "🔎 Finding Prometheus datasource UID..."
DATASOURCE_UID=$(curl -s -u "$GRAFANA_USER:$GRAFANA_PASS" "$GRAFANA_URL/api/datasources/name/$DATASOURCE_NAME" | jq -r '.uid')

if [ -z "$DATASOURCE_UID" ] || [ "$DATASOURCE_UID" == "null" ]; then
    echo "❌ Could not find Prometheus datasource named '$DATASOURCE_NAME' in Grafana."
    exit 1
fi
echo "✅ Found Prometheus datasource UID: $DATASOURCE_UID"

echo "📋 Importing dashboard from: $DASHBOARD_PATH"

# Read the dashboard file and create the final JSON payload using jq
DASHBOARD_JSON=$(cat "$DASHBOARD_PATH")
PAYLOAD=$(jq -n \
  --argjson dashboard "$DASHBOARD_JSON" \
  --arg ds_uid "$DATASOURCE_UID" \
  '{
    dashboard: $dashboard,
    overwrite: true,
    inputs: [
      {
        name: "DS_PROMETHEUS",
        type: "datasource",
        pluginId: "prometheus",
        value: $ds_uid
      }
    ]
  }')

RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -u "$GRAFANA_USER:$GRAFANA_PASS" \
    -d "$PAYLOAD" \
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