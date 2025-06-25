# Social Media Tags Testing Guide

## ✅ Current Status

Your social media meta tags are **correctly implemented** and generating properly. Here's what's working:

### ✅ Working Tags
- ✅ Open Graph (Facebook/LinkedIn): `og:type`, `og:title`, `og:description`, `og:url`, `og:site_name`
- ✅ Twitter Cards: `twitter:card`, `twitter:title`, `twitter:description`, `twitter:image`, `twitter:creator`, `twitter:site`
- ✅ Article metadata: `article:author`, `article:published_time`, `article:tag` (10 tags found)
- ✅ Proper Twitter handle format: `@lancerogers`

## ⚠️ Missing Default Image

**Issue**: The default OG image referenced in `site.yml` doesn't exist:
```
default_image: "/static/images/blockhead-og-default.png"
```

**File location**: `/Users/lancerogers/Dev/BlockheadConsulting/Website/static/images/blockhead-og-default.png`

## 🔧 Pre-Ship Testing Checklist

### 1. **Fix Missing Image** (Required)
Create or add the default OG image:
```bash
# Create a 1200x630 image for optimal social media display
# Place it at: static/images/blockhead-og-default.png
```

**Recommended dimensions**: 1200x630 pixels (Facebook's recommended size)

### 2. **Local Testing** (Quick validation)
```bash
# Run the provided test script
./scripts/test-social-tags.sh

# Or test manually
curl -s http://localhost:8087/blog/1_million_per_week_with_claude_code | grep -E "(og:|twitter:)"
```

### 3. **Production URL Testing** (Essential before shipping)

Once deployed to production, test with these validators:

#### Facebook/Meta Debugger
- **URL**: https://developers.facebook.com/tools/debug/
- **Test URL**: `https://blockheadconsulting.com/blog/[your-post-slug]`
- **What it checks**: Open Graph tags, image loading, description formatting
- **Fix**: Use "Scrape Again" button if you make changes

#### Twitter Card Validator  
- **URL**: https://cards-dev.twitter.com/validator
- **Test URL**: `https://blockheadconsulting.com/blog/[your-post-slug]`
- **What it checks**: Twitter Card format, image display, character limits
- **Note**: May require Twitter Developer account

#### LinkedIn Post Inspector
- **URL**: https://www.linkedin.com/post-inspector/
- **Test URL**: `https://blockheadconsulting.com/blog/[your-post-slug]`
- **What it checks**: Professional network display, image quality

### 4. **Manual Social Media Testing** (Real-world validation)

#### Test on Facebook
1. Create a test post with your blog URL
2. Check if title, description, and image appear correctly
3. Verify the link preview looks professional

#### Test on Twitter
1. Tweet your blog URL
2. Verify the card shows: title, description, image, and author
3. Check that the image is crisp and readable

#### Test on LinkedIn
1. Share your blog URL in a LinkedIn post
2. Verify professional appearance
3. Check that your name appears as the author

### 5. **Different Blog Posts Testing**
Test multiple blog posts to ensure consistency:
```bash
# Test different posts
./scripts/test-social-tags.sh # Modify script to test different URLs
```

### 6. **Image Optimization Checklist**
- [ ] Image exists and is accessible (no 404)
- [ ] Optimal dimensions: 1200x630 pixels
- [ ] File size: < 5MB (preferably < 1MB)
- [ ] Format: PNG or JPG
- [ ] Clear text/logo that's readable when resized
- [ ] Matches your brand colors/style

## 🚀 Deployment Testing Steps

1. **Deploy to staging/production**
2. **Wait 15 minutes** (for CDN/cache propagation)
3. **Test each validator above**
4. **Share on personal social accounts** (test posts)
5. **Check analytics** for social media referral traffic

## 🔍 Common Issues to Watch For

### Facebook/Meta Issues
- Image not loading → Check image accessibility
- Old data showing → Use "Scrape Again" in debugger
- Description truncated → Keep under 160 characters

### Twitter Issues  
- Card not showing → Verify `twitter:card` type
- Image cropped badly → Use 1200x630 dimensions
- Author missing → Check `twitter:creator` tag

### LinkedIn Issues
- Professional appearance → Use business-appropriate images
- Author attribution → Verify `article:author` tag
- Company branding → Check `og:site_name`

## 📊 Success Metrics

After deployment, monitor:
- **Social media referral traffic** in analytics
- **Click-through rates** from social platforms
- **Engagement rates** on shared posts
- **Brand recognition** from consistent image/messaging

## 🛠️ Quick Fixes

If issues arise after deployment:

1. **Image problems**: 
   - Check image URL accessibility
   - Verify dimensions and file size
   - Use Facebook debugger "Scrape Again"

2. **Text problems**:
   - Verify meta tag content in page source
   - Check for HTML encoding issues
   - Ensure descriptions are under character limits

3. **Cache issues**:
   - Wait 24 hours for full propagation
   - Use "Scrape Again" or similar tools
   - Clear CDN cache if applicable