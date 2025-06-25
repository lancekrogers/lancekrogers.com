import { debug } from '../utils/debug.js';

// Track animation state
let initialAnimationsRun = false;
let heroAnimationRunning = false;

// Export getters for state
export function hasInitialAnimationsRun() {
  return initialAnimationsRun;
}

export function isHeroAnimationRunning() {
  return heroAnimationRunning;
}

// Initialize animations function
export function initializeAnimations() {
  // Only run on home page
  const isHomePage = window.location.pathname === '/' || 
                     document.querySelector('.hero') !== null;
  
  if (!isHomePage) return;
  
  debug("Initializing animations - isHomePage:", isHomePage, "initialAnimationsRun:", initialAnimationsRun);
  
  // Only run full animation sequence on initial page load
  if (!initialAnimationsRun) {
    // No more global glitch effects - just initialize typing
    initializeHeroTextAnimation(true);
    // Set flag AFTER the animation is set up properly
    setTimeout(() => {
      initialAnimationsRun = true;
    }, 100);
  } else {
    // Just show static text with blinking cursor on navigation back
    initializeHeroTextAnimation(false);
    
    // Always ensure subtitle is visible when navigating back to home
    const heroSubtitle = document.querySelector('.hero-subtitle');
    if (heroSubtitle) {
      debug("Ensuring subtitle is visible on navigation back");
      heroSubtitle.classList.add('fade-in');
      heroSubtitle.style.opacity = '1';
    }
  }
}

export function initializeHeroTextAnimation(shouldType = true) {
  debug("initializeHeroTextAnimation called with shouldType:", shouldType, "initialAnimationsRun:", initialAnimationsRun);
  
  // Prevent multiple animation loops
  if (heroAnimationRunning && shouldType) {
    debug("Animation already running, skipping");
    return;
  }
  
  // Terminal typewriter effect setup
  const glitchElement = document.querySelector(".glitch");
  // In professional mode, we use the glitch element itself as the terminal content
  const terminalContent = glitchElement;
  const heroSubtitle = document.querySelector(".hero-subtitle");
  
  if (!glitchElement || !terminalContent) return;
  
  // Get hero style configuration
  const heroStyle = glitchElement.getAttribute('data-hero-style') || 'professional';
  
  // Reset any existing content and state
  glitchElement.classList.remove("subtle-glitch");
  
  // Reset visibility for fresh animations only on first load
  if (shouldType && !initialAnimationsRun) {
    if (heroSubtitle) heroSubtitle.classList.remove('fade-in');
  }
  
  heroAnimationRunning = true;
  
  // Apply hero style enhancements and get calculated delay
  const calculatedDelay = applyHeroEnhancements(heroStyle, glitchElement, terminalContent, shouldType);
  
  // The text to type
  const fullText = "BLOCKHEAD CONSULTING";
  
  if (!shouldType) {
    // Just show static text with NO cursor (navigation back to home)
    terminalContent.textContent = fullText;
    glitchElement.setAttribute('data-text', fullText);
    
    // Show subtitle immediately when navigating back
    const heroSubtitle = document.querySelector('.hero-subtitle');
    
    if (heroSubtitle) {
      heroSubtitle.classList.add('fade-in');
      heroSubtitle.style.opacity = '1';
      heroSubtitle.style.transition = 'none';
    }
    
    // Set up periodic glitches for static text
    setupPeriodicGlitches(glitchElement);
    heroAnimationRunning = false; // Reset flag for navigation
    return;
  }
  
  // In professional mode, the boot sequence and fade-in is handled by applyProfessionalEffects
  // Do NOT return early - let the timing complete properly
  if (heroStyle === 'professional' && shouldType) {
    heroAnimationRunning = false; // Allow future animations
    // applyProfessionalEffects has already set up the boot sequence and fade-in timing
    return;
  }
  
  // Original typing animation for initial load
  terminalContent.textContent = "";
  glitchElement.setAttribute('data-text', ""); // Start with empty data-text for glitch effects
  
  
  let i = 0;
  let animationStartTime = performance.now() + calculatedDelay;
  let lastTypeTime = 0;
  let typingComplete = false;
  let typingStarted = false;
  
  function animationLoop() {
    const now = performance.now();
    
    // Calculate dynamic typing delay based on position and context
    let typingDelay = 120; // Base delay
    
    if (!typingComplete && i < fullText.length) {
      const currentChar = fullText[i];
      const isSpace = currentChar === ' ';
      const currentWord = fullText.substring(0, i + 1);
      
      // Calculate context-specific delays
      if (isSpace) {
        typingDelay = 300 + Math.random() * 100; // 300-400ms pause between words
      } else if (currentWord.endsWith('BLOCK')) {
        typingDelay = 420 + Math.random() * 100; // 420-520ms pause after "BLOCK"
      } else if (currentWord.endsWith('HEAD')) {
        typingDelay = 620 + Math.random() * 200; // 620-820ms pause after "HEAD"
      } else {
        const rand = Math.random();
        if (rand < 0.1) {
          // Occasional longer pause (like thinking)
          typingDelay = 160 + Math.random() * 80; // 160-240ms
        } else if (rand < 0.3) {
          // Brief hesitation
          typingDelay = 130 + Math.random() * 30; // 130-160ms
        } else {
          // Normal typing
          typingDelay = 120; // Base speed
        }
      }
    }
    
    // Handle clean terminal typewriter effect
    if (!typingComplete && now >= animationStartTime && now - lastTypeTime >= typingDelay) {
      // Show cursor and start blinking when typing begins
      if (!typingStarted) {
        typingStarted = true;
      }
      
      if (i < fullText.length) {
        const currentText = fullText.substring(0, i + 1);
        terminalContent.textContent = currentText;
        glitchElement.setAttribute('data-text', currentText); // Sync data-text with visible text
        
        i++;
        lastTypeTime = now;
      } else {
        typingComplete = true;
        // Set final data-text attribute for future glitch effects
        glitchElement.setAttribute('data-text', fullText);
        // Hide cursor immediately when typing completes
        
        // Fade out CRT effects in cyberpunk mode after typing completes
        const heroSection = document.querySelector('.hero');
        if (heroSection && heroSection.classList.contains('hero-cyberpunk')) {
          setTimeout(() => {
            heroSection.classList.add('crt-fade-out');
          }, 2000); // Wait 2 seconds after typing completes
        }
        
        // Fade in the subtitle after typing completes (cyberpunk mode only)
        // In professional mode, this is handled by applyProfessionalEffects after boot sequence
        if (heroStyle !== 'professional') {
          setTimeout(() => {
            const heroSubtitle = document.querySelector('.hero-subtitle');
            if (heroSubtitle) {
              heroSubtitle.classList.add('fade-in');
            }
          }, 800); // Delay after typing completes
        }
      }
    }
    
    // Only continue the loop if the element still exists and we're on the home page
    if (document.querySelector('.glitch') && 
        (window.location.pathname === '/' || document.querySelector('.hero'))) {
      requestAnimationFrame(animationLoop);
    }
  }

  // Start the animation loop
  requestAnimationFrame(animationLoop);
}

export function setupPeriodicGlitches(glitchElement) {
  // No implementation needed - removed periodic glitches
  debug('Periodic glitches disabled');
}

export function applyHeroEnhancements(heroStyle, glitchElement, terminalContent, shouldType) {
  const heroSection = document.querySelector('.hero');
  
  if (heroStyle === 'professional') {
    return applyProfessionalEffects(glitchElement, terminalContent, shouldType);
  } else if (heroStyle === 'cyberpunk' && heroSection) {
    heroSection.classList.add('hero-cyberpunk');
    return applyCyberpunkEffects(glitchElement, terminalContent, shouldType);
  }
  
  // Default delay
  return 0;
}

export function applyProfessionalEffects(glitchElement, terminalContent, shouldType) {
  debug("Applying professional effects with shouldType:", shouldType);
  
  // Professional mode: Show boot sequence effect
  if (shouldType) {
    // Set the full text immediately for professional mode
    const fullText = "BLOCKHEAD CONSULTING";
    terminalContent.textContent = fullText;
    glitchElement.setAttribute('data-text', fullText);
    
    // Start boot sequence using the already loaded function
    debug("Starting boot sequence");
    if (window.createBootSequence) {
      // Always run boot sequence when navigating to homepage
      const bootDuration = window.createBootSequence('professional', terminalContent);
      debug("Boot sequence started, duration:", bootDuration);
      
      // Show subtitle after boot sequence completes
      setTimeout(() => {
        const heroSubtitle = document.querySelector('.hero-subtitle');
        if (heroSubtitle) {
          debug("Showing subtitle after boot sequence");
          heroSubtitle.classList.remove('fade-in'); // Remove class first
          void heroSubtitle.offsetWidth; // Force reflow
          heroSubtitle.classList.add('fade-in'); // Re-add for animation
        }
      }, bootDuration + 500);
    } else {
      debug("createBootSequence not found, showing subtitle immediately");
      // Fallback: show subtitle after short delay
      setTimeout(() => {
        const heroSubtitle = document.querySelector('.hero-subtitle');
        if (heroSubtitle) {
          heroSubtitle.classList.remove('fade-in');
          void heroSubtitle.offsetWidth;
          heroSubtitle.classList.add('fade-in');
        }
      }, 2000);
    }
    
    // Return long delay to prevent normal typing animation from interfering
    return 999999; 
  }
  
  // For navigation back, no delay needed
  return 0;
}

export function applyCyberpunkEffects(glitchElement, terminalContent, shouldType) {
  debug("Applying cyberpunk effects");
  
  // Add base animations
  const animations = [];
  
  // Digital rain or crypto-AI background
  const bgChoice = Math.random();
  if (bgChoice < 0.5) {
    createMatrixRain();
    animations.push('matrix-rain');
  } else {
    createAICryptoBackground();
    animations.push('ai-crypto');
  }
  
  // Standard delay for cyberpunk mode
  return 500; // Start typing after half a second
}

export function createMatrixRain() {
  debug('Matrix rain effect would be created here');
  // Implementation moved to blockchain-animations.js (disabled)
}

export function createAICryptoBackground() {
  debug('AI/Crypto background effect would be created here');
  // Implementation moved to blockchain-animations.js (disabled)
}

// Reset animation state (useful for testing)
export function resetAnimationState() {
  initialAnimationsRun = false;
  heroAnimationRunning = false;
}