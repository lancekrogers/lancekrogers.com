/**
 * End-to-End Test Suite
 * 
 * This file tests the actual navigation issues reported:
 * 1. "Learn more" link goes to bottom of about page
 * 2. "Back to blog" goes to home page instead of blog
 * 3. Hamburger menu breaks after navigation
 * 4. Tagline missing after navigation
 * 5. Footer blog link appends /blog to URL
 * 6. Contact button not working
 */

const { describe, it, expect, beforeAll, afterAll } = require('@jest/globals');
const puppeteer = require('puppeteer');

describe('E2E Navigation Tests', () => {
  let browser;
  let page;
  const baseUrl = 'http://localhost:8087';

  beforeAll(async () => {
    browser = await puppeteer.launch({
      headless: true,
      args: ['--no-sandbox', '--disable-setuid-sandbox']
    });
    page = await browser.newPage();
    
    // Set mobile viewport for some tests
    await page.setViewport({ width: 375, height: 812 });
  });

  afterAll(async () => {
    await browser.close();
  });

  describe('Bio "Learn More" Link', () => {
    it('should navigate to top of about page', async () => {
      await page.goto(baseUrl);
      
      // Wait for bio section
      await page.waitForSelector('.bio-link');
      
      // Click learn more
      await page.click('.bio-link');
      
      // Wait for navigation
      await page.waitForFunction(() => window.location.pathname === '/about');
      
      // Check scroll position
      const scrollY = await page.evaluate(() => window.scrollY);
      expect(scrollY).toBe(0); // Should be at top
    });
  });

  describe('Blog "Back to Blog" Button', () => {
    it('should navigate to blog page, not home', async () => {
      // Navigate to a blog post
      await page.goto(`${baseUrl}/blog/test-post`);
      
      // Wait for back button
      await page.waitForSelector('.back-link');
      
      // Click back to blog
      await page.click('.back-link');
      
      // Wait for navigation
      await page.waitForFunction(() => window.location.pathname === '/blog');
      
      // Verify we're on blog page
      const url = await page.url();
      expect(url).toBe(`${baseUrl}/blog`);
    });
  });

  describe('Mobile Hamburger Menu', () => {
    it('should work after navigation', async () => {
      // Start on home page
      await page.goto(baseUrl);
      
      // Test initial hamburger
      await page.waitForSelector('#hamburger-toggle');
      await page.click('#hamburger-toggle');
      
      let menuActive = await page.$eval('#mobile-menu', el => 
        el.classList.contains('active')
      );
      expect(menuActive).toBe(true);
      
      // Navigate to about
      await page.click('.mobile-menu a[href="/about"]');
      await page.waitForFunction(() => window.location.pathname === '/about');
      
      // Navigate back to home
      await page.click('.mobile-menu a[href="/"]');
      await page.waitForFunction(() => window.location.pathname === '/');
      
      // Test hamburger still works
      await page.waitForSelector('#hamburger-toggle');
      await page.click('#hamburger-toggle');
      
      menuActive = await page.$eval('#mobile-menu', el => 
        el.classList.contains('active')
      );
      expect(menuActive).toBe(true);
    });
  });

  describe('Hero Tagline Visibility', () => {
    it('should show tagline after navigation back to home', async () => {
      // Navigate away and back
      await page.goto(`${baseUrl}/about`);
      await page.goto(baseUrl);
      
      // Wait for tagline
      await page.waitForSelector('.hero-subtitle', { visible: true });
      
      // Check opacity
      const opacity = await page.$eval('.hero-subtitle', el => 
        window.getComputedStyle(el).opacity
      );
      expect(opacity).toBe('1');
    });
  });

  describe('Footer Navigation', () => {
    it('should navigate correctly without URL issues', async () => {
      await page.goto(baseUrl);
      
      // Scroll to footer
      await page.evaluate(() => {
        document.querySelector('footer').scrollIntoView();
      });
      
      // Click footer blog link
      await page.waitForSelector('footer a[href="/blog"]');
      await page.click('footer a[href="/blog"]');
      
      // Wait for navigation
      await page.waitForFunction(() => window.location.pathname === '/blog');
      
      // Check URL is correct
      const url = await page.url();
      expect(url).toBe(`${baseUrl}/blog`);
    });
  });

  describe('Contact Button', () => {
    it('should scroll to contact section', async () => {
      await page.goto(baseUrl);
      
      // Click contact link
      await page.click('a[href="#contact"]');
      
      // Wait for smooth scroll
      await page.waitForTimeout(1000);
      
      // Check if contact section is in view
      const contactInView = await page.evaluate(() => {
        const contact = document.querySelector('#contact');
        const rect = contact.getBoundingClientRect();
        return rect.top >= 0 && rect.top <= window.innerHeight;
      });
      
      expect(contactInView).toBe(true);
    });
  });

  describe('Complete Navigation Flow', () => {
    it('should maintain state through full navigation', async () => {
      const checkState = async () => {
        // Check tagline
        const taglineVisible = await page.$eval('.hero-subtitle', el => 
          window.getComputedStyle(el).opacity === '1'
        ).catch(() => false);
        
        // Check buttons
        const buttonsVisible = await page.$eval('.hero-cta', el => 
          window.getComputedStyle(el).opacity === '1'
        ).catch(() => false);
        
        return { taglineVisible, buttonsVisible };
      };
      
      // Start on home
      await page.goto(baseUrl);
      let state = await checkState();
      expect(state.taglineVisible).toBe(true);
      expect(state.buttonsVisible).toBe(true);
      
      // Navigate to about
      await page.click('a[href="/about"]');
      await page.waitForFunction(() => window.location.pathname === '/about');
      
      // Navigate to blog
      await page.click('a[href="/blog"]');
      await page.waitForFunction(() => window.location.pathname === '/blog');
      
      // Navigate back to home
      await page.click('a[href="/"]');
      await page.waitForFunction(() => window.location.pathname === '/');
      
      // Check state is maintained
      state = await checkState();
      expect(state.taglineVisible).toBe(true);
      expect(state.buttonsVisible).toBe(true);
    });
  });

  describe('Boot Sequence on Mobile', () => {
    it('should fade out properly on mobile', async () => {
      await page.goto(baseUrl);
      
      // Wait for boot sequence to appear
      const bootExists = await page.waitForSelector('.boot-sequence', {
        timeout: 2000
      }).catch(() => null);
      
      if (bootExists) {
        // Wait for it to fade out
        await page.waitForFunction(() => {
          const boot = document.querySelector('.boot-sequence');
          return !boot || window.getComputedStyle(boot).opacity === '0';
        }, { timeout: 15000 });
        
        // Verify it's removed from DOM
        await page.waitForFunction(() => 
          !document.querySelector('.boot-sequence')
        , { timeout: 2000 });
      }
      
      // Boot sequence should be gone
      const bootGone = await page.$('.boot-sequence');
      expect(bootGone).toBe(null);
    });
  });
});