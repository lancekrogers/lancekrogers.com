#!/bin/bash
# Script to clean up console logging for production deployment

echo "🧹 Cleaning up console logging for production..."

# Backup original files
echo "📦 Creating backups..."
mkdir -p backups/js
cp -r static/js backups/js/
cp static/*.js backups/

# Remove test mode file from production
echo "🗑️  Removing test mode script..."
rm -f static/mode-tests.js

# Clean up main-init.js
echo "🔧 Cleaning main-init.js..."
sed -i '' '/console\.log.*About button clicked/d' static/js/main-init.js
sed -i '' '/console\.log.*Click debounced/d' static/js/main-init.js
sed -i '' '/console\.log.*HTMX content swap/d' static/js/main-init.js
sed -i '' '/console\.log.*Request:/d' static/js/main-init.js
sed -i '' '/console\.warn.*Cleaning up stuck request/d' static/js/main-init.js

# Clean up blog-enhanced.js
echo "🔧 Cleaning blog-enhanced.js..."
sed -i '' '/console\.log.*Blog enhanced features initialized/d' static/js/blog-enhanced.js
sed -i '' '/console\.log.*Search query:/d' static/js/blog-enhanced.js
sed -i '' '/console\.log.*Pagination:/d' static/js/blog-enhanced.js
sed -i '' '/console\.log.*Filter:/d' static/js/blog-enhanced.js
sed -i '' '/console\.log.*Tag clicked:/d' static/js/blog-enhanced.js

# Clean up boot-sequence.js
echo "🔧 Cleaning boot-sequence.js..."
sed -i '' '/console\.log.*Boot sequence/d' static/boot-sequence.js
sed -i '' '/console\.log.*Starting typewriter/d' static/boot-sequence.js
sed -i '' '/console\.log.*Glitch animation/d' static/boot-sequence.js
sed -i '' '/console\.log.*DOM fully loaded/d' static/boot-sequence.js

# Clean up blog-post.js
echo "🔧 Cleaning blog-post.js..."
sed -i '' '/console\.error.*Mermaid rendering failed/d' static/js/blog-post.js
sed -i '' '/console\.error.*Failed to fetch/d' static/js/blog-post.js

# Clean up popups.js
echo "🔧 Cleaning popups.js..."
sed -i '' '/console\.log.*CLOSE BUTTON CLICKED!/d' static/js/modules/popups.js
sed -i '' '/console\.log.*Modal state/d' static/js/modules/popups.js
sed -i '' '/console\.log.*Dialog debug/d' static/js/modules/popups.js

# Set debugLogging to false in templates
echo "🔧 Setting debugLogging to false..."
sed -i '' 's/debugLogging = {{\.Config\.ConsoleLogging}}/debugLogging = false/' templates/layouts/base.html
sed -i '' 's/debugLogging = {{\.Config\.ConsoleLogging}}/debugLogging = false/' templates/layouts/base-blog-post.html

# Clean up Go DEBUG logging
echo "🔧 Cleaning Go debug logs..."
sed -i '' '/log\.Printf("DEBUG:/d' internal/handlers/blog.go

echo "✅ Console logging cleaned for production!"
echo "📋 Summary of changes:"
echo "   - Removed mode-tests.js"
echo "   - Removed console.log statements from JS files"
echo "   - Set debugLogging to false"
echo "   - Removed DEBUG logs from Go files"
echo ""
echo "⚠️  Backups created in ./backups directory"
echo "💡 To restore: cp -r backups/js/* static/js/"