#!/bin/bash

# Final Social Media Tags Test
echo "🎯 Final Social Media Tags Validation"
echo "====================================="
echo ""

# Test the blog post URL
URL="http://localhost:8087/blog/1_million_per_week_with_claude_code"
echo "Testing: $URL"
echo ""

# Get the HTML
HTML=$(curl -s "$URL")

# Test key meta tags
echo "🔍 Key Meta Tags:"
echo "=================="

echo "$HTML" | grep -E "(og:title|og:description|og:image|og:url)" | while read line; do
    echo "✅ $line"
done

echo ""
echo "$HTML" | grep -E "(twitter:card|twitter:title|twitter:image|twitter:creator)" | while read line; do
    echo "✅ $line"
done

echo ""
echo "🖼️  Image Accessibility Test:"
echo "============================"

# Test the OG image
OG_IMAGE=$(echo "$HTML" | grep -o 'property="og:image" content="[^"]*"' | grep -o 'https://[^"]*')
if [ ! -z "$OG_IMAGE" ]; then
    STATUS=$(curl -o /dev/null -s -w "%{http_code}" "$OG_IMAGE")
    echo "Image URL: $OG_IMAGE"
    echo "HTTP Status: $STATUS"
    if [ "$STATUS" = "200" ]; then
        echo "✅ Default OG image is accessible!"
    else
        echo "❌ Image returns HTTP $STATUS"
    fi
else
    echo "❌ No OG image found"
fi

echo ""
echo "🚀 Ready for Production Testing!"
echo ""
echo "Next steps:"
echo "1. Deploy to production"
echo "2. Test with Facebook Debugger: https://developers.facebook.com/tools/debug/"
echo "3. Test with Twitter Card Validator: https://cards-dev.twitter.com/validator"
echo "4. Test with LinkedIn Post Inspector: https://www.linkedin.com/post-inspector/"
echo "5. Share on personal social accounts to verify real-world appearance"