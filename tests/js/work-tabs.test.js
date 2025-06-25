const { JSDOM } = require('jsdom');

describe('Work Tabs Functionality', () => {
  let dom;
  let document;
  let window;

  beforeEach(() => {
    // Create DOM with work tabs HTML structure
    const html = `
      <!DOCTYPE html>
      <html>
        <body>
          <!-- Mobile Tab Navigation -->
          <div class="work-tabs-mobile">
            <button class="tab-btn active" data-tab="ai">AI/ML</button>
            <button class="tab-btn" data-tab="crypto">Blockchain</button>
            <button class="tab-btn" data-tab="fintech">FinTech</button>
          </div>

          <!-- Mobile Tab Content -->
          <div class="work-tabs-content-mobile">
            <div class="tab-content active" data-tab-content="ai">
              <div class="work-row clickable" data-work-item="ai-guild-framework">
                <span class="row-name">Guild Framework</span>
              </div>
            </div>
            <div class="tab-content" data-tab-content="crypto">
              <div class="work-row clickable" data-work-item="crypto-mythical-games">
                <span class="row-name">Mythical Games</span>
              </div>
            </div>
            <div class="tab-content" data-tab-content="fintech">
              <div class="work-row clickable" data-work-item="fintech-bank-of-america">
                <span class="row-name">Bank of America</span>
              </div>
            </div>
          </div>

          <!-- Desktop Containers -->
          <div class="work-containers-desktop">
            <div class="work-container" data-sector="ai">
              <div class="work-row clickable" data-work-item="ai-guild-framework">
                <span class="row-name">Guild Framework</span>
              </div>
            </div>
            <div class="work-container" data-sector="crypto">
              <div class="work-row clickable" data-work-item="crypto-mythical-games">
                <span class="row-name">Mythical Games</span>
              </div>
            </div>
          </div>

          <!-- Work Popups -->
          <div class="work-popup" data-work-popup="ai-guild-framework">
            <div class="work-popup-content">
              <button class="work-popup-close">&times;</button>
              <h3>Guild Framework</h3>
            </div>
          </div>
        </body>
      </html>
    `;

    dom = new JSDOM(html, { 
      pretendToBeVisual: true,
      resources: "usable" 
    });
    
    document = dom.window.document;
    window = dom.window;
    
    global.document = document;
    global.window = window;
  });

  afterEach(() => {
    if (dom) {
      dom.window.close();
    }
  });

  describe('HTML Structure', () => {
    it('should have mobile tab navigation', () => {
      const tabsContainer = document.querySelector('.work-tabs-mobile');
      const tabs = document.querySelectorAll('.tab-btn');
      
      expect(tabsContainer).not.toBeNull();
      expect(tabs.length).toBe(3);
      expect(tabs[0].dataset.tab).toBe('ai');
      expect(tabs[1].dataset.tab).toBe('crypto');
      expect(tabs[2].dataset.tab).toBe('fintech');
    });

    it('should have mobile tab content sections', () => {
      const contentSections = document.querySelectorAll('.tab-content');
      
      expect(contentSections.length).toBe(3);
      expect(contentSections[0].dataset.tabContent).toBe('ai');
      expect(contentSections[1].dataset.tabContent).toBe('crypto');
      expect(contentSections[2].dataset.tabContent).toBe('fintech');
    });

    it('should have default active tab set to AI', () => {
      const activeTab = document.querySelector('.tab-btn.active');
      const activeContent = document.querySelector('.tab-content.active');
      
      expect(activeTab).not.toBeNull();
      expect(activeTab.dataset.tab).toBe('ai');
      expect(activeContent).not.toBeNull();
      expect(activeContent.dataset.tabContent).toBe('ai');
    });
  });

  describe('Desktop Containers', () => {
    it('should have work containers for desktop view', () => {
      const aiContainer = document.querySelector('[data-sector="ai"]');
      const cryptoContainer = document.querySelector('[data-sector="crypto"]');
      
      expect(aiContainer).not.toBeNull();
      expect(cryptoContainer).not.toBeNull();
    });

    it('should have clickable work rows', () => {
      const workRows = document.querySelectorAll('.work-row.clickable');
      
      expect(workRows.length).toBeGreaterThan(0);
      
      // Check that rows have proper data attributes
      workRows.forEach(row => {
        expect(row.dataset.workItem).toBeTruthy();
      });
    });
  });

  describe('Work Popups', () => {
    it('should have popup elements for work items', () => {
      const popup = document.querySelector('[data-work-popup="ai-guild-framework"]');
      
      expect(popup).not.toBeNull();
      expect(popup.querySelector('.work-popup-content')).not.toBeNull();
      expect(popup.querySelector('.work-popup-close')).not.toBeNull();
    });

    it('should have close button in popups', () => {
      const popup = document.querySelector('[data-work-popup="ai-guild-framework"]');
      const closeBtn = popup.querySelector('.work-popup-close');
      
      expect(closeBtn).not.toBeNull();
      expect(closeBtn.textContent).toBe('×');
    });
  });

  describe('SVG Icons', () => {
    it('should use text instead of emoji in tab buttons', () => {
      const tabButtons = document.querySelectorAll('.tab-btn');
      
      tabButtons.forEach(btn => {
        const text = btn.textContent;
        // Check for common emoji patterns
        expect(text).not.toMatch(/[\u{1F300}-\u{1F9FF}]/u);
        expect(text).not.toMatch(/🤖|💼|🏦/);
      });
    });
  });

  describe('Work Row Structure', () => {
    it('should have proper structure for work rows', () => {
      const workRow = document.querySelector('[data-work-item="ai-guild-framework"]');
      const rowName = workRow.querySelector('.row-name');
      
      expect(workRow).not.toBeNull();
      expect(rowName).not.toBeNull();
      expect(rowName.textContent).toBe('Guild Framework');
    });

    it('should have work rows in both mobile and desktop views', () => {
      const mobileRow = document.querySelector('.tab-content [data-work-item="ai-guild-framework"]');
      const desktopRow = document.querySelector('.work-containers-desktop [data-work-item="ai-guild-framework"]');
      
      expect(mobileRow).not.toBeNull();
      expect(desktopRow).not.toBeNull();
    });
  });
});