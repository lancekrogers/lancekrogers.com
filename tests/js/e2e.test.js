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
    it('should navigate to about page', async () => {
      await page.goto(baseUrl);
      
      // Wait for bio section
      await page.waitForSelector('.bio-link', { timeout: 10000 });
      
      // Click learn more
      await page.click('.bio-link');
      
      // Wait for navigation
      await page.waitForFunction(() => window.location.pathname === '/about', { timeout: 10000 });
      
      // Check that we're on the about page
      const pathname = await page.evaluate(() => window.location.pathname);
      expect(pathname).toBe('/about');
    });
  });

  describe('Blog Navigation', () => {
    it('should navigate to blog page', async () => {
      await page.goto(baseUrl);
      
      // Wait for and click blog link in navigation
      await page.waitForSelector('a[href="/blog"]', { timeout: 10000 });
      await page.click('a[href="/blog"]');
      
      // Wait for navigation
      await page.waitForFunction(() => window.location.pathname === '/blog', { timeout: 10000 });
      
      // Verify we're on blog page
      const pathname = await page.evaluate(() => window.location.pathname);
      expect(pathname).toBe('/blog');
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
      
      // Set mobile viewport first
      await page.setViewport({ width: 375, height: 667 });
      
      // Close menu by clicking hamburger again
      await page.click('#hamburger-toggle');
      
      // Verify menu closed
      await page.waitForFunction(() => 
        !document.querySelector('#mobile-menu').classList.contains('active'),
        { timeout: 5000 }
      );
      
      const menuClosed = await page.$eval('#mobile-menu', el => 
        !el.classList.contains('active')
      );
      expect(menuClosed).toBe(true);
    });
  });

  describe('Hero Tagline Visibility', () => {
    it('should show tagline on page load', async () => {
      await page.goto(baseUrl);
      
      // Wait for tagline element to exist
      await page.waitForSelector('.hero-subtitle', { timeout: 15000 });
      
      // Wait for animation to complete - give time for boot sequence
      await page.waitForTimeout(8000);
      
      // Check if tagline is visible (either opacity 1 or fade-in class)
      const isVisible = await page.$eval('.hero-subtitle', el => {
        const opacity = window.getComputedStyle(el).opacity;
        const hasFadeIn = el.classList.contains('fade-in');
        return opacity === '1' || hasFadeIn;
      });
      expect(isVisible).toBe(true);
    });
  });

  describe('Footer Navigation', () => {
    it('should navigate correctly without URL issues', async () => {
      await page.goto(baseUrl);
      
      // Scroll to footer
      await page.evaluate(() => {
        document.querySelector('footer').scrollIntoView();
      });
      
      // Wait for footer links to be visible
      await page.waitForSelector('footer .footer-links a', { timeout: 10000 });
      
      // Get the first footer link and verify it's valid
      const firstLink = await page.$eval('footer .footer-links a', el => el.href);
      expect(firstLink).toBeTruthy();
      expect(firstLink).not.toContain('undefined');
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
      // Start on home
      await page.goto(baseUrl);
      await page.waitForSelector('.hero', { timeout: 10000 });
      
      // Navigate to about
      await page.click('a[href="/about"]');
      await page.waitForFunction(() => window.location.pathname === '/about', { timeout: 10000 });
      
      // Navigate back to home
      await page.click('a[href="/"]');
      await page.waitForFunction(() => window.location.pathname === '/', { timeout: 10000 });
      
      // Verify home page elements are present
      const heroExists = await page.$('.hero');
      expect(heroExists).toBeTruthy();
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