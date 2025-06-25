#!/bin/bash
# Fix blog posts not showing in production

echo "🔧 Fixing production blog setup..."

# Add SQLite configuration to production environment
echo "📝 Adding SQLite configuration to .env.production.example..."
cat >> .env.production.example << 'EOF'

# Blog Configuration
BLOG_ENABLED=true
BLOG_SQLITE_ENABLED=true
BLOG_DB_PATH=./data/blog.db
EOF

# Create production setup script
echo "📝 Creating production setup script..."
cat > scripts/setup_production_data.sh << 'EOF'
#!/bin/bash
# Setup production data directories and database

echo "📁 Creating data directories..."
mkdir -p data/messages
mkdir -p data/mermaid-temp
mkdir -p data/cache

echo "🔧 Setting permissions..."
chmod 755 data
chmod 755 data/messages
chmod 755 data/mermaid-temp
chmod 755 data/cache

echo "✅ Production data directories ready!"
echo ""
echo "📋 Next steps:"
echo "1. Ensure BLOG_SQLITE_ENABLED=true in your .env file"
echo "2. The blog database will be created automatically on first run"
echo "3. Blog posts will be imported from embedded files"
EOF

chmod +x scripts/setup_production_data.sh

# Update Makefile prod-deploy target
echo "🔧 Updating Makefile for production data setup..."
cat > scripts/makefile_prod_patch.txt << 'EOF'
# Add this after line 697 in the prod-deploy target
	@echo "   └─ Setting up production data directories..."
	@ssh -p $(PROD_SSH_PORT) $(PROD_USER)@$(PROD_HOST) "cd $(PROD_PATH) && mkdir -p data/messages data/mermaid-temp data/cache && chmod -R 755 data"
EOF

echo "✅ Fix script complete!"
echo ""
echo "📋 To fix production:"
echo "1. On your Oracle server, add to .env:"
echo "   BLOG_ENABLED=true"
echo "   BLOG_SQLITE_ENABLED=true"
echo "   BLOG_DB_PATH=./data/blog.db"
echo ""
echo "2. Create data directories:"
echo "   ssh to server and run:"
echo "   cd /home/lance/Blockhead.Consulting"
echo "   mkdir -p data/messages data/mermaid-temp data/cache"
echo "   chmod -R 755 data"
echo ""
echo "3. Restart the service:"
echo "   sudo systemctl restart blockhead"
echo ""
echo "The blog database will be created and populated on first run."