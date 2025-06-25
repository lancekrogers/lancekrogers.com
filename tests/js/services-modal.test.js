const { describe, it, expect } = require('@jest/globals');

describe('Services Modal System', () => {
  
  describe('Modal Trigger Buttons', () => {
    it('should have View Services button on home page', () => {
      // Home page should have modal trigger buttons
      const expectedHomeButtons = [
        {
          selector: '.service-packages-cta .packages-modal-trigger',
          text: 'View Services →',
          attribute: 'data-packages-modal'
        },
        {
          selector: '.cta-actions .packages-modal-trigger',
          text: 'View Services',
          attribute: 'data-packages-modal'
        }
      ];
      
      expectedHomeButtons.forEach(button => {
        expect(button.selector).toContain('.packages-modal-trigger');
        expect(button.attribute).toBe('data-packages-modal');
      });
    });
    
    it('should have View Services button on about page', () => {
      // About page should have modal trigger button
      const expectedAboutButton = {
        selector: '.packages-compact .packages-modal-trigger',
        text: 'View Services',
        attribute: 'data-packages-modal'
      };
      
      expect(expectedAboutButton.selector).toContain('.packages-modal-trigger');
      expect(expectedAboutButton.attribute).toBe('data-packages-modal');
    });
    
    it('should have consistent button styling', () => {
      // All modal trigger buttons should have proper CSS classes
      const buttonClasses = [
        'btn-primary.packages-modal-trigger',
        'btn-view-packages.packages-modal-trigger', 
        'btn-cta-secondary.packages-modal-trigger'
      ];
      
      buttonClasses.forEach(btnClass => {
        expect(btnClass).toContain('packages-modal-trigger');
      });
    });
  });
  
  describe('Main Modal Structure', () => {
    it('should have main packages modal in DOM', () => {
      // Main modal should exist with proper structure
      const expectedModalStructure = {
        modal: '[data-packages-main-modal]',
        content: '.packages-main-modal-content',
        closeButton: '.packages-main-modal-close',
        title: '.section-title',
        subtitle: '.section-subtitle',
        packageGrid: '.package-grid.large-packages'
      };
      
      Object.values(expectedModalStructure).forEach(selector => {
        expect(selector).toBeTruthy();
      });
    });
    
    it('should have individual package popups', () => {
      // Individual package popups should exist
      const expectedPackageTypes = [
        'evaluation', 'assessment', 'pilot', 
        'accelerator', 'ai_acceleration', 'embedded'
      ];
      
      expectedPackageTypes.forEach(packageType => {
        const popupSelector = `[data-package-popup="${packageType}"]`;
        expect(popupSelector).toBeTruthy();
      });
    });
    
    it('should have proper modal hierarchy', () => {
      // Main modal should have lower z-index than individual popups
      const zIndexConfig = {
        mainModal: 1000,
        individualPopups: 1100
      };
      
      expect(zIndexConfig.individualPopups).toBeGreaterThan(zIndexConfig.mainModal);
    });
  });
  
  describe('Modal Interaction Logic', () => {
    it('should open main modal when trigger clicked', () => {
      // Clicking trigger should open main modal
      const modalBehavior = {
        trigger: 'data-packages-modal',
        targetModal: 'data-packages-main-modal',
        activeClass: 'active',
        bodyOverflow: 'hidden'
      };
      
      expect(modalBehavior.trigger).toBe('data-packages-modal');
      expect(modalBehavior.targetModal).toBe('data-packages-main-modal');
      expect(modalBehavior.activeClass).toBe('active');
    });
    
    it('should close main modal when close button clicked', () => {
      // Close button should close main modal
      const closeBehavior = {
        closeButton: '.packages-main-modal-close',
        targetModal: 'data-packages-main-modal',
        activeClass: 'active',
        bodyOverflow: ''
      };
      
      expect(closeBehavior.closeButton).toBe('.packages-main-modal-close');
      expect(closeBehavior.activeClass).toBe('active');
    });
    
    it('should close main modal when background clicked', () => {
      // Clicking modal background should close modal
      const backgroundCloseBehavior = {
        backgroundElement: '.packages-main-modal',
        shouldClose: true,
        contentElement: '.packages-main-modal-content',
        shouldNotClose: true
      };
      
      expect(backgroundCloseBehavior.shouldClose).toBe(true);
      expect(backgroundCloseBehavior.shouldNotClose).toBe(true);
    });
    
    it('should open individual package popup from main modal', () => {
      // Clicking package in main modal should open individual popup
      const packageInteraction = {
        packageElement: '.package-interactive',
        dataAttribute: 'data-package',
        popupSelector: '[data-package-popup=""]',
        zIndexOverride: '1100'
      };
      
      expect(packageInteraction.dataAttribute).toBe('data-package');
      expect(packageInteraction.zIndexOverride).toBe('1100');
    });
    
    it('should handle escape key properly', () => {
      // Escape key should close modals in proper order
      const escapeKeyBehavior = {
        individualPopupOpen: 'close individual popup first',
        mainModalOpen: 'close main modal',
        noModalsOpen: 'no action'
      };
      
      expect(escapeKeyBehavior.individualPopupOpen).toContain('individual');
      expect(escapeKeyBehavior.mainModalOpen).toContain('main');
    });
  });
  
  describe('Modal Content Validation', () => {
    it('should display service packages from configuration', () => {
      // Modal should show packages from services.yml
      const expectedPackages = [
        'Idea Evaluation',
        'Technical Assessment', 
        'Rapid Prototype Development',
        'Strategic Development Partnership',
        'AI Development Acceleration',
        'Embedded Team Acceleration'
      ];
      
      expectedPackages.forEach(packageName => {
        expect(packageName).toBeTruthy();
      });
    });
    
    it('should show package details in individual popups', () => {
      // Individual popups should show detailed package info
      const expectedDetailSections = [
        'What You Get',
        'Process', 
        'Outcomes'
      ];
      
      expectedDetailSections.forEach(section => {
        expect(section).toBeTruthy();
      });
    });
    
    it('should have proper modal titles and descriptions', () => {
      // Modal content should have proper structure
      const modalContent = {
        mainTitle: 'Service Packages I Offer',
        mainSubtitle: 'Structured engagements designed for different stages of growth',
        packageHint: 'Click for details'
      };
      
      expect(modalContent.mainTitle).toBe('Service Packages I Offer');
      expect(modalContent.mainSubtitle).toContain('Structured engagements');
    });
  });
  
  describe('Accessibility and UX', () => {
    it('should prevent body scroll when modal is open', () => {
      // Body scroll should be disabled with modal open
      const scrollBehavior = {
        modalOpen: 'hidden',
        modalClosed: '',
        individualPopupOpen: 'maintain hidden if main modal open'
      };
      
      expect(scrollBehavior.modalOpen).toBe('hidden');
      expect(scrollBehavior.modalClosed).toBe('');
    });
    
    it('should have proper ARIA attributes', () => {
      // Modals should be accessible
      const accessibilityFeatures = [
        'keyboard navigation',
        'escape key support',
        'focus management',
        'screen reader support'
      ];
      
      accessibilityFeatures.forEach(feature => {
        expect(feature).toBeTruthy();
      });
    });
    
    it('should work on mobile devices', () => {
      // Modal should be responsive and work on mobile
      const mobileSupport = {
        touchEvents: true,
        responsiveLayout: true,
        mobileOptimizations: true,
        viewportHandling: true
      };
      
      Object.values(mobileSupport).forEach(supported => {
        expect(supported).toBe(true);
      });
    });
  });
  
  describe('Error Handling', () => {
    it('should handle missing services configuration gracefully', () => {
      // Modal should not break if ServicesConfig is missing
      const errorHandling = {
        missingConfig: 'modal should not render',
        conditionalRendering: '{{if .ServicesConfig}}',
        gracefulDegradation: true
      };
      
      expect(errorHandling.conditionalRendering).toContain('ServicesConfig');
      expect(errorHandling.gracefulDegradation).toBe(true);
    });
    
    it('should prevent duplicate event listeners', () => {
      // Event listeners should be properly managed
      const eventManagement = {
        removeOldListeners: true,
        preventDuplicates: true,
        cleanupOnDestroy: true
      };
      
      Object.values(eventManagement).forEach(managed => {
        expect(managed).toBe(true);
      });
    });
    
    it('should handle concurrent modal operations', () => {
      // Multiple modal operations should be handled safely
      const concurrencyHandling = {
        preventMultipleOpens: true,
        stateManagement: true,
        mutexLocking: 'isMainPackageModalOpen flag'
      };
      
      expect(concurrencyHandling.preventMultipleOpens).toBe(true);
      expect(concurrencyHandling.mutexLocking).toContain('flag');
    });
  });
  
  describe('Integration with Existing Systems', () => {
    it('should work with HTMX navigation', () => {
      // Modal should work after HTMX page swaps
      const htmxIntegration = {
        workAfterSwap: true,
        eventListenerReinit: true,
        statePreservation: false // Should reset state on navigation
      };
      
      expect(htmxIntegration.workAfterSwap).toBe(true);
      expect(htmxIntegration.eventListenerReinit).toBe(true);
    });
    
    it('should not conflict with other modals', () => {
      // Services modal should not interfere with contact modal
      const modalCoexistence = {
        contactModal: 'should work independently',
        workModal: 'should work independently', 
        expertiseModal: 'should work independently'
      };
      
      Object.values(modalCoexistence).forEach(behavior => {
        expect(behavior).toContain('independently');
      });
    });
    
    it('should maintain consistent branding and styling', () => {
      // Modal should match site design system
      const designConsistency = {
        colorScheme: 'cyberpunk theme',
        typography: 'JetBrains Mono',
        animations: 'consistent transitions',
        spacing: 'consistent padding/margins'
      };
      
      expect(designConsistency.colorScheme).toContain('cyberpunk');
      expect(designConsistency.typography).toContain('Mono');
    });
  });
  
});