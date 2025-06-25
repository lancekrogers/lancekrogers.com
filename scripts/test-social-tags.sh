#!/bin/bash

# Social Media Tags Validator Script
# Tests Open Graph and Twitter Card meta tags

set -e

# Configuration
BASE_URL="http://localhost:8087"
BLOG_POST="1_million_per_week_with_claude_code"
EXPECTED_SITE_NAME="Blockhead Consulting"
EXPECTED_TWITTER="@lancerogers"

echo "🔍 Testing Social Media Meta Tags"
echo "================================="
echo ""

# Test URL
TEST_URL="${BASE_URL}/blog/${BLOG_POST}"
echo "Testing URL: $TEST_URL"
echo ""

# Fetch the HTML
HTML=$(curl -s "$TEST_URL")

# Test required Open Graph tags
echo "📘 Open Graph Tags:"
echo "==================="

# Check og:type
OG_TYPE=$(echo "$HTML" | grep -o 'property="og:type" content="[^"]*"' | grep -o 'content="[^"]*"' | cut -d'"' -f2)
echo "og:type: $OG_TYPE"
if [ "$OG_TYPE" = "article" ]; then
    echo "✅ og:type is correct"
else
    echo "❌ og:type should be 'article', got '$OG_TYPE'"
fi

# Check og:title
OG_TITLE=$(echo "$HTML" | grep -o 'property="og:title" content="[^"]*"' | grep -o 'content="[^"]*"' | cut -d'"' -f2)
echo "og:title: $OG_TITLE"
if [ ! -z "$OG_TITLE" ]; then
    echo "✅ og:title is present"
else
    echo "❌ og:title is missing"
fi

# Check og:description
OG_DESC=$(echo "$HTML" | grep -o 'property="og:description" content="[^"]*"' | grep -o 'content="[^"]*"' | cut -d'"' -f2 | head -1)
echo "og:description: ${OG_DESC:0:80}..."
if [ ! -z "$OG_DESC" ]; then
    echo "✅ og:description is present"
else
    echo "❌ og:description is missing"
fi

# Check og:url
OG_URL=$(echo "$HTML" | grep -o 'property="og:url" content="[^"]*"' | grep -o 'content="[^"]*"' | cut -d'"' -f2)
echo "og:url: $OG_URL"
if [[ "$OG_URL" == *"/blog/$BLOG_POST" ]]; then
    echo "✅ og:url is correct"
else
    echo "❌ og:url should contain '/blog/$BLOG_POST'"
fi

# Check og:image
OG_IMAGE=$(echo "$HTML" | grep -o 'property="og:image" content="[^"]*"' | grep -o 'content="[^"]*"' | cut -d'"' -f2)
echo "og:image: $OG_IMAGE"
if [ ! -z "$OG_IMAGE" ]; then
    echo "✅ og:image is present"
else
    echo "❌ og:image is missing"
fi

# Check og:site_name
OG_SITE=$(echo "$HTML" | grep -o 'property="og:site_name" content="[^"]*"' | grep -o 'content="[^"]*"' | cut -d'"' -f2)
echo "og:site_name: $OG_SITE"
if [ "$OG_SITE" = "$EXPECTED_SITE_NAME" ]; then
    echo "✅ og:site_name is correct"
else
    echo "❌ og:site_name should be '$EXPECTED_SITE_NAME', got '$OG_SITE'"
fi

echo ""

# Test Twitter Card tags
echo "🐦 Twitter Card Tags:"
echo "===================="

# Check twitter:card
TWITTER_CARD=$(echo "$HTML" | grep -o 'name="twitter:card" content="[^"]*"' | grep -o 'content="[^"]*"' | cut -d'"' -f2)
echo "twitter:card: $TWITTER_CARD"
if [ "$TWITTER_CARD" = "summary_large_image" ]; then
    echo "✅ twitter:card is correct"
else
    echo "❌ twitter:card should be 'summary_large_image', got '$TWITTER_CARD'"
fi

# Check twitter:title
TWITTER_TITLE=$(echo "$HTML" | grep -o 'name="twitter:title" content="[^"]*"' | grep -o 'content="[^"]*"' | cut -d'"' -f2)
echo "twitter:title: $TWITTER_TITLE"
if [ ! -z "$TWITTER_TITLE" ]; then
    echo "✅ twitter:title is present"
else
    echo "❌ twitter:title is missing"
fi

# Check twitter:description
TWITTER_DESC=$(echo "$HTML" | grep -o 'name="twitter:description" content="[^"]*"' | grep -o 'content="[^"]*"' | cut -d'"' -f2 | head -1)
echo "twitter:description: ${TWITTER_DESC:0:80}..."
if [ ! -z "$TWITTER_DESC" ]; then
    echo "✅ twitter:description is present"
else
    echo "❌ twitter:description is missing"
fi

# Check twitter:image
TWITTER_IMAGE=$(echo "$HTML" | grep -o 'name="twitter:image" content="[^"]*"' | grep -o 'content="[^"]*"' | cut -d'"' -f2)
echo "twitter:image: $TWITTER_IMAGE"
if [ ! -z "$TWITTER_IMAGE" ]; then
    echo "✅ twitter:image is present"
else
    echo "❌ twitter:image is missing"
fi

# Check twitter:creator
TWITTER_CREATOR=$(echo "$HTML" | grep -o 'name="twitter:creator" content="[^"]*"' | grep -o 'content="[^"]*"' | cut -d'"' -f2)
echo "twitter:creator: $TWITTER_CREATOR"
if [ "$TWITTER_CREATOR" = "$EXPECTED_TWITTER" ]; then
    echo "✅ twitter:creator is correct"
else
    echo "❌ twitter:creator should be '$EXPECTED_TWITTER', got '$TWITTER_CREATOR'"
fi

# Check twitter:site
TWITTER_SITE=$(echo "$HTML" | grep -o 'name="twitter:site" content="[^"]*"' | grep -o 'content="[^"]*"' | cut -d'"' -f2)
echo "twitter:site: $TWITTER_SITE"
if [ "$TWITTER_SITE" = "$EXPECTED_TWITTER" ]; then
    echo "✅ twitter:site is correct"
else
    echo "❌ twitter:site should be '$EXPECTED_TWITTER', got '$TWITTER_SITE'"
fi

echo ""

# Test Article tags
echo "📄 Article Meta Tags:"
echo "===================="

# Check article:author
ARTICLE_AUTHOR=$(echo "$HTML" | grep -o 'property="article:author" content="[^"]*"' | grep -o 'content="[^"]*"' | cut -d'"' -f2)
echo "article:author: $ARTICLE_AUTHOR"
if [ ! -z "$ARTICLE_AUTHOR" ]; then
    echo "✅ article:author is present"
else
    echo "❌ article:author is missing"
fi

# Check article:published_time
ARTICLE_PUBLISHED=$(echo "$HTML" | grep -o 'property="article:published_time" content="[^"]*"' | grep -o 'content="[^"]*"' | cut -d'"' -f2)
echo "article:published_time: $ARTICLE_PUBLISHED"
if [ ! -z "$ARTICLE_PUBLISHED" ]; then
    echo "✅ article:published_time is present"
else
    echo "❌ article:published_time is missing"
fi

# Count article:tag
TAG_COUNT=$(echo "$HTML" | grep -c 'property="article:tag"' || echo "0")
echo "article:tag count: $TAG_COUNT"
if [ "$TAG_COUNT" -gt 0 ]; then
    echo "✅ article:tag tags are present ($TAG_COUNT found)"
else
    echo "❌ article:tag tags are missing"
fi

echo ""

# Test image accessibility
echo "🖼️  Image Validation:"
echo "===================="

if [ ! -z "$OG_IMAGE" ]; then
    HTTP_STATUS=$(curl -o /dev/null -s -w "%{http_code}" "$OG_IMAGE")
    echo "Image URL: $OG_IMAGE"
    echo "HTTP Status: $HTTP_STATUS"
    if [ "$HTTP_STATUS" = "200" ]; then
        echo "✅ Image is accessible"
    else
        echo "❌ Image is not accessible (HTTP $HTTP_STATUS)"
    fi
else
    echo "❌ No image URL to test"
fi

echo ""
echo "🎉 Social Media Tag Validation Complete!"
echo ""
echo "💡 Next Steps:"
echo "   1. Test with production URL on Facebook Debugger"
echo "   2. Test with Twitter Card Validator"
echo "   3. Check LinkedIn Post Inspector"
echo "   4. Test different blog posts to ensure consistency"