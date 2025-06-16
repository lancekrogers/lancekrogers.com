// Conditional console logging based on configuration
let debugLogging = false; // Will be set from server config

function debug(...args) {
  if (debugLogging) {
    console.log(...args);
  }
}

function debugError(...args) {
  if (debugLogging) {
    console.error(...args);
  }
}

// Enhanced smooth scroll for anchor links and services navigation
function initializeSmoothScroll() {
  // Handle direct anchor links (like #services when already on home page)  
  document.querySelectorAll('a[href^="#"]:not([href*="/"])').forEach((anchor) => {
    anchor.addEventListener("click", function (e) {
      e.preventDefault();
      const targetId = this.getAttribute("href");
      const target = document.querySelector(targetId);
      if (target) {
        // Add a small delay to ensure any ongoing animations don't interfere
        setTimeout(() => {
          target.scrollIntoView({
            behavior: "smooth",
            block: "start",
          });
        }, 100);
      }
    });
  });

  // Handle services links that need to load home page first
  document.querySelectorAll('a[data-scroll-to="services"]').forEach((servicesLink) => {
    servicesLink.addEventListener("click", function (e) {
      // If we're already on the home page, just scroll to services
      if (window.location.pathname === '/' && document.getElementById('services')) {
        e.preventDefault();
        document.getElementById('services').scrollIntoView({
          behavior: "smooth",
          block: "start",
        });
      }
      // Otherwise, let HTMX handle it and the scroll will happen in htmx:afterSwap
    });
  });
}

// Initialize on page load
initializeSmoothScroll();

// Mobile hamburger menu functionality
let mobileMenuInitialized = false;

function initializeMobileMenu() {
  const hamburgerToggle = document.getElementById('hamburger-toggle');
  const mobileMenu = document.getElementById('mobile-menu');
  
  if (hamburgerToggle && mobileMenu && !mobileMenuInitialized) {
    // Reset menu state first
    hamburgerToggle.classList.remove('active');
    mobileMenu.classList.remove('active');
    
    // Add hamburger toggle event listener
    hamburgerToggle.addEventListener('click', function() {
      hamburgerToggle.classList.toggle('active');
      mobileMenu.classList.toggle('active');
    });

    // Close menu when clicking on any mobile menu link
    mobileMenu.addEventListener('click', function(e) {
      if (e.target.tagName === 'A') {
        hamburgerToggle.classList.remove('active');
        mobileMenu.classList.remove('active');
      }
    });
    
    mobileMenuInitialized = true;
  }
}

// Global click handler for closing mobile menu (only attach once)
let globalMobileHandlerAttached = false;

function attachGlobalMobileMenuHandler() {
  if (!globalMobileHandlerAttached) {
    document.addEventListener('click', function(event) {
      const hamburgerToggle = document.getElementById('hamburger-toggle');
      const mobileMenu = document.getElementById('mobile-menu');
      
      if (hamburgerToggle && mobileMenu) {
        if (!hamburgerToggle.contains(event.target) && !mobileMenu.contains(event.target)) {
          hamburgerToggle.classList.remove('active');
          mobileMenu.classList.remove('active');
        }
      }
    });
    
    globalMobileHandlerAttached = true;
  }
}

// Initialize mobile menu
initializeMobileMenu();
attachGlobalMobileMenuHandler();

// Add active class to current nav item and hide home link on home page
function updateNavigation() {
  const currentLocation = location.pathname;
  const desktopMenuItems = document.querySelectorAll(".desktop-nav a");
  const mobileMenuItems = document.querySelectorAll(".mobile-menu a");
  const desktopHomeLink = document.querySelector(".desktop-nav a[href='/']");
  const mobileHomeLink = document.querySelector(".mobile-menu a[href='/']");
  
  // Update desktop navigation
  desktopMenuItems.forEach((item) => {
    item.classList.remove("active");
    if (item.getAttribute("href") === currentLocation) {
      item.classList.add("active");
    }
  });
  
  // Update mobile navigation
  mobileMenuItems.forEach((item) => {
    item.classList.remove("active");
    if (item.getAttribute("href") === currentLocation) {
      item.classList.add("active");
    }
  });
  
  // Hide home link when on home page
  const isHomePage = currentLocation === '/' || 
                     document.querySelector('.hero') !== null ||
                     document.querySelector('.glitch') !== null;
  
  if (desktopHomeLink) {
    if (isHomePage) {
      desktopHomeLink.style.display = 'none';
    } else {
      desktopHomeLink.style.display = '';
    }
  }
  
  if (mobileHomeLink) {
    if (isHomePage) {
      mobileHomeLink.style.display = 'none';
    } else {
      mobileHomeLink.style.display = '';
    }
  }
}

// Run on initial load
updateNavigation();

// Track if initial animations have run
let initialAnimationsRun = false;
let heroAnimationRunning = false;

// Initialize animations function
function initializeAnimations() {
  // Only run on home page
  const isHomePage = window.location.pathname === '/' || 
                     document.querySelector('.hero') !== null;
  
  if (!isHomePage) return;
  
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
  }
}

function initializeHeroTextAnimation(shouldType = true) {
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
  const terminalCursor = document.querySelector(".terminal-cursor") || { style: {} }; // Provide fallback object
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
  const calculatedDelay = applyHeroEnhancements(heroStyle, glitchElement, terminalContent, terminalCursor, shouldType);
  
  // The text to type
  const fullText = "BLOCKHEAD CONSULTING";
  
  if (!shouldType) {
    // Just show static text with NO cursor (navigation back to home)
    terminalContent.textContent = fullText;
    glitchElement.setAttribute('data-text', fullText);
    terminalCursor.style.display = 'none'; // Hide cursor completely
    
    // Show subtitle immediately when navigating back
    const heroSubtitle = document.querySelector('.hero-subtitle');
    
    if (heroSubtitle) {
      heroSubtitle.classList.add('fade-in');
      heroSubtitle.style.opacity = '1';
      heroSubtitle.style.transition = 'none';
    }
    
    // Set up periodic glitches for static text
    setupPeriodicGlitches(glitchElement, terminalCursor);
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
  
  // Hide cursor initially (during boot sequence)
  terminalCursor.style.display = 'none';
  
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
        terminalCursor.style.display = 'inline';
        terminalCursor.style.animation = 'cursor-blink 1s infinite';
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
        terminalCursor.style.display = 'none';
        
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
    
    // No more periodic glitch effects - remove this functionality
    
    // Only continue the loop if the element still exists and we're on the home page
    if (document.querySelector('.glitch') && 
        (window.location.pathname === '/' || document.querySelector('.hero'))) {
      requestAnimationFrame(animationLoop);
    }
  }

  // Start the animation loop
  requestAnimationFrame(animationLoop);
}

// Package popup functionality
let isPopupOpen = false;

function initializePackagePopups() {
  // Remove any existing event listeners first
  document.removeEventListener('click', handlePackageClicks);
  
  // Add global click handler for package interactions
  document.addEventListener('click', handlePackageClicks);
}

function handlePackageClicks(e) {
  // Handle package popup close button clicks first
  if (e.target.classList.contains('package-popup-close')) {
    e.preventDefault();
    e.stopPropagation();
    hideAllPackagePopups();
    return;
  }
  
  // Handle expertise popup close button clicks
  if (e.target.classList.contains('expertise-popup-close')) {
    e.preventDefault();
    e.stopPropagation();
    hideAllExpertisePopups();
    return;
  }
  
  // Handle package popup background clicks to close
  if (e.target.classList.contains('package-popup')) {
    e.preventDefault();
    e.stopPropagation();
    hideAllPackagePopups();
    return;
  }
  
  // Handle expertise popup background clicks to close
  if (e.target.classList.contains('expertise-popup')) {
    e.preventDefault();
    e.stopPropagation();
    hideAllExpertisePopups();
    return;
  }
  
  // Don't handle clicks inside popup content
  if (e.target.closest('.package-popup-content') || e.target.closest('.expertise-popup-content')) {
    return;
  }
  
  // Handle package card clicks
  if (e.target.closest('.package-interactive')) {
    const packageCard = e.target.closest('.package-interactive');
    const packageId = packageCard.dataset.package;
    const popup = document.querySelector(`[data-package-popup="${packageId}"]`);
    
    if (popup) {
      e.preventDefault();
      e.stopPropagation();
      showPackagePopup(popup);
    }
  }
  
  // Handle expertise card clicks
  if (e.target.closest('.expertise-interactive')) {
    const expertiseCard = e.target.closest('.expertise-interactive');
    const expertiseId = expertiseCard.dataset.expertise;
    const popup = document.querySelector(`[data-expertise-popup="${expertiseId}"]`);
    
    if (popup) {
      e.preventDefault();
      e.stopPropagation();
      showExpertisePopup(popup);
    }
  }
}

function showPackagePopup(popup) {
  // Prevent multiple popups from opening
  if (isPopupOpen) return;
  
  // Hide any other open popups first
  hideAllPackagePopups();
  
  // Show the selected popup
  popup.classList.add('active');
  isPopupOpen = true;
  
  // Prevent body scroll when popup is open
  document.body.style.overflow = 'hidden';
}

function hideAllPackagePopups() {
  const allPopups = document.querySelectorAll('.package-popup');
  allPopups.forEach(popup => {
    popup.classList.remove('active');
  });
  
  // Reset flag
  isPopupOpen = false;
  
  // Restore body scroll
  document.body.style.overflow = '';
}

function showExpertisePopup(popup) {
  // Prevent multiple popups from opening
  if (isPopupOpen) return;
  
  // Hide any other open popups first
  hideAllExpertisePopups();
  hideAllPackagePopups();
  
  // Show the selected popup
  popup.classList.add('active');
  isPopupOpen = true;
  
  // Prevent body scroll when popup is open
  document.body.style.overflow = 'hidden';
}

function hideAllExpertisePopups() {
  const allPopups = document.querySelectorAll('.expertise-popup');
  allPopups.forEach(popup => {
    popup.classList.remove('active');
  });
  
  // Reset flag
  isPopupOpen = false;
  
  // Restore body scroll
  document.body.style.overflow = '';
}

// Work popup functionality
let isWorkPopupOpen = false;

function initializeWorkPopups() {
  // Remove any existing event listeners first
  document.removeEventListener('click', handleWorkClicks);
  
  // Add global click handler for work interactions
  document.addEventListener('click', handleWorkClicks);
}

function handleWorkClicks(e) {
  // Handle work popup close button clicks
  if (e.target.classList.contains('work-popup-close')) {
    e.preventDefault();
    e.stopPropagation();
    hideAllWorkPopups();
    return;
  }
  
  // Handle work popup background clicks to close
  if (e.target.classList.contains('work-popup')) {
    e.preventDefault();
    e.stopPropagation();
    hideAllWorkPopups();
    return;
  }
  
  // Don't handle clicks inside popup content
  if (e.target.closest('.work-popup-content')) {
    return;
  }
  
  // Handle work item clicks
  if (e.target.closest('.work-item.interactive')) {
    const workItem = e.target.closest('.work-item.interactive');
    const workItemId = workItem.dataset.workItem;
    const popup = document.querySelector(`[data-work-popup="${workItemId}"]`);
    
    if (popup) {
      e.preventDefault();
      e.stopPropagation();
      showWorkPopup(popup);
    }
  }
}

function showWorkPopup(popup) {
  // Prevent multiple popups from opening
  if (isWorkPopupOpen) return;
  
  // Hide any other open popups first
  hideAllWorkPopups();
  hideAllPackagePopups();
  hideAllExpertisePopups();
  
  // Show the selected popup
  popup.classList.add('active');
  isWorkPopupOpen = true;
  
  // Prevent body scroll when popup is open
  document.body.style.overflow = 'hidden';
}

function hideAllWorkPopups() {
  const allPopups = document.querySelectorAll('.work-popup');
  allPopups.forEach(popup => {
    popup.classList.remove('active');
  });
  
  // Reset flag
  isWorkPopupOpen = false;
  
  // Restore body scroll
  document.body.style.overflow = '';
}

// Handle escape key to close popups
document.addEventListener('keydown', function(e) {
  if (e.key === 'Escape') {
    hideAllPackagePopups();
    hideAllExpertisePopups();
    hideAllWorkPopups();
  }
});


// Run on initial page load
document.addEventListener('DOMContentLoaded', function() {
  initializeAnimations();
  initializePackagePopups();
  initializeWorkPopups();
  initializeFormEncryption();
  
  // Fallback: Ensure subtitle is visible after 8 seconds if animations fail
  setTimeout(() => {
    const heroSubtitle = document.querySelector('.hero-subtitle');
    
    if (heroSubtitle && window.getComputedStyle(heroSubtitle).opacity === '0') {
      heroSubtitle.style.opacity = '1';
      heroSubtitle.style.transition = 'opacity 0.5s ease-out';
    }
  }, 8000);
  
  // Initialize calendar if we're on the calendar page
  if (window.location.pathname === '/calendar') {
    initializeCalendar();
  }
});

// Calendar functionality
let currentWeekOffset = 0;
let availableSlots = [];
let selectedSlot = null;

async function loadSlots() {
  try {
    const response = await fetch("/api/slots");
    availableSlots = await response.json();
    displaySlots();
  } catch (error) {
    debugError("Error loading slots:", error);
    const slotsElement = document.getElementById("time-slots");
    if (slotsElement) {
      slotsElement.innerHTML = '<p class="error">Error loading available times. Please try again.</p>';
    }
  }
}

function displaySlots() {
  const container = document.getElementById("time-slots");
  if (!container) return;
  
  const startDate = new Date();
  startDate.setDate(startDate.getDate() + currentWeekOffset * 7);

  const endDate = new Date(startDate);
  endDate.setDate(endDate.getDate() + 7);

  // Filter slots for current week
  const weekSlots = availableSlots.filter((slot) => {
    const slotDate = new Date(slot.date);
    return slotDate >= startDate && slotDate < endDate;
  });

  if (weekSlots.length === 0) {
    container.innerHTML = '<p class="no-slots">No available times this week. Try another week.</p>';
    return;
  }

  // Group by date
  const slotsByDate = {};
  weekSlots.forEach((slot) => {
    if (!slotsByDate[slot.date]) {
      slotsByDate[slot.date] = [];
    }
    slotsByDate[slot.date].push(slot);
  });

  // Display
  container.innerHTML = "";
  Object.entries(slotsByDate).forEach(([date, slots]) => {
    const dateDiv = document.createElement("div");
    dateDiv.className = "date-group";

    const dateObj = new Date(date + "T00:00:00");
    const dayName = dateObj.toLocaleDateString("en-US", { weekday: "short" });
    const monthDay = dateObj.toLocaleDateString("en-US", { month: "short", day: "numeric" });

    dateDiv.innerHTML = `
      <div class="date-header">
        <div class="day-name">${dayName}</div>
        <div class="month-day">${monthDay}</div>
      </div>
      <div class="time-slots">
        ${slots.map(slot => `
          <button class="time-slot" data-slot-id="${slot.id}" data-date="${slot.date}" data-time="${slot.time}">
            ${slot.time}
          </button>
        `).join("")}
      </div>
    `;

    container.appendChild(dateDiv);
  });

  // Add click handlers
  document.querySelectorAll(".time-slot").forEach((button) => {
    button.addEventListener("click", selectSlot);
  });

  // Update week display
  updateWeekDisplay(startDate, endDate);
}

function updateWeekDisplay(start, end) {
  const display = document.getElementById("current-week");
  if (!display) return;
  
  const startStr = start.toLocaleDateString("en-US", { month: "short", day: "numeric" });
  const endStr = end.toLocaleDateString("en-US", { month: "short", day: "numeric" });
  display.textContent = `${startStr} - ${endStr}`;
}

function selectSlot(e) {
  const button = e.target;
  selectedSlot = {
    id: button.dataset.slotId,
    date: button.dataset.date,
    time: button.dataset.time,
  };

  // Update UI
  document.querySelectorAll(".time-slot").forEach((btn) => btn.classList.remove("selected"));
  button.classList.add("selected");

  // Show booking form
  showBookingForm();
}

function showBookingForm() {
  const calendarContainer = document.getElementById("calendar-container");
  const bookingForm = document.getElementById("booking-form");
  
  if (calendarContainer && bookingForm) {
    calendarContainer.classList.add("hidden");
    bookingForm.classList.remove("hidden");

    // Display selected time
    const dateObj = new Date(selectedSlot.date + "T00:00:00");
    const dateStr = dateObj.toLocaleDateString("en-US", {
      weekday: "long",
      month: "long",
      day: "numeric",
    });
    
    const selectedTimeElement = document.getElementById("selected-time");
    const selectedSlotElement = document.getElementById("selected-slot");
    
    if (selectedTimeElement) {
      selectedTimeElement.textContent = `${dateStr} at ${selectedSlot.time} MT`;
    }
    if (selectedSlotElement) {
      selectedSlotElement.value = selectedSlot.id;
    }
  }
}

function showCalendar() {
  const calendarContainer = document.getElementById("calendar-container");
  const bookingForm = document.getElementById("booking-form");
  
  if (calendarContainer && bookingForm) {
    bookingForm.classList.add("hidden");
    calendarContainer.classList.remove("hidden");
  }
}

function initializeCalendar() {
  // Add event handlers for calendar navigation
  const prevWeekBtn = document.getElementById("prev-week");
  const nextWeekBtn = document.getElementById("next-week");
  const cancelBookingBtn = document.getElementById("cancel-booking");
  const bookingForm = document.getElementById("booking-details");

  if (prevWeekBtn) {
    prevWeekBtn.addEventListener("click", () => {
      currentWeekOffset--;
      displaySlots();
    });
  }

  if (nextWeekBtn) {
    nextWeekBtn.addEventListener("click", () => {
      currentWeekOffset++;
      displaySlots();
    });
  }

  if (cancelBookingBtn) {
    cancelBookingBtn.addEventListener("click", showCalendar);
  }

  if (bookingForm) {
    bookingForm.addEventListener("submit", async (e) => {
      e.preventDefault();

      const formData = new FormData(e.target);
      const data = Object.fromEntries(formData);

      try {
        const response = await fetch("/api/book", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(data),
        });

        const result = await response.json();

        if (response.ok) {
          // Show confirmation
          const bookingFormEl = document.getElementById("booking-form");
          const confirmationEl = document.getElementById("confirmation");
          
          if (bookingFormEl && confirmationEl) {
            bookingFormEl.classList.add("hidden");
            confirmationEl.classList.remove("hidden");

            const dateObj = new Date(selectedSlot.date + "T00:00:00");
            const dateStr = dateObj.toLocaleDateString("en-US", {
              weekday: "long",
              month: "long",
              day: "numeric",
            });
            
            const confirmationDetails = document.getElementById("confirmation-details");
            if (confirmationDetails) {
              confirmationDetails.textContent = `${dateStr} at ${selectedSlot.time} MT`;
            }
          }
        } else {
          const bookingResponse = document.getElementById("booking-response");
          if (bookingResponse) {
            bookingResponse.innerHTML = `<div class="alert error">${result.message || "Booking failed. Please try again."}</div>`;
          }
        }
      } catch (error) {
        debugError("Booking error:", error);
        const bookingResponse = document.getElementById("booking-response");
        if (bookingResponse) {
          bookingResponse.innerHTML = '<div class="alert error">An error occurred. Please try again.</div>';
        }
      }
    });
  }

  // Load slots
  loadSlots();
}

// Track Services button clicks for HTMX navigation
let servicesBtnClickForScroll = false;

// Detect Services button clicks before HTMX request
document.addEventListener('htmx:beforeRequest', function(evt) {
  if (evt.detail.elt && evt.detail.elt.getAttribute('data-scroll-to') === 'services') {
    servicesBtnClickForScroll = true;
  } else {
    servicesBtnClickForScroll = false;
  }
});

// Run when HTMX loads new content
document.addEventListener('htmx:afterSwap', function(evt) {
  // Update navigation state for all content swaps (with small delay to ensure DOM is updated)
  setTimeout(updateNavigation, 10);
  
  // Re-initialize smooth scroll for new content
  setTimeout(initializeSmoothScroll, 50);
  
  // Re-initialize package popups for new content
  setTimeout(initializePackagePopups, 50);
  
  // Re-initialize work popups for new content
  setTimeout(initializeWorkPopups, 50);
  
  // Re-initialize form encryption for new content
  setTimeout(initializeFormEncryption, 50);
  
  // Mobile menu is initialized once globally, no need to re-initialize
  
  // Scroll to top for about page navigation
  if (evt.detail.target.id === 'main-content' && 
      evt.detail.xhr.responseURL.includes('/content/about')) {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }
  
  // Scroll to top for work page navigation
  if (evt.detail.target.id === 'main-content' && 
      evt.detail.xhr.responseURL.includes('/content/work')) {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }
  
  // Scroll to top for blog page navigation
  if (evt.detail.target.id === 'main-content' && 
      evt.detail.xhr.responseURL.includes('/content/blog')) {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }
  
  // Initialize calendar for calendar page content
  if (evt.detail.target.id === 'main-content' && 
      evt.detail.xhr.responseURL.includes('/content/calendar')) {
    initializeCalendar();
  }
  
  // Initialize hero text animation for home page content (no animations on navigation)
  if (evt.detail.target.id === 'main-content' && 
      evt.detail.xhr.responseURL.includes('/content/home')) {
    // Small delay to ensure DOM is fully loaded
    setTimeout(() => {
      // Always show static text when navigating back to home
      initializeHeroTextAnimation(false);
    }, 50);
    
    // Services scroll fix: Check if this was triggered by clicking Services button
    if (servicesBtnClickForScroll && evt.detail.xhr.responseURL.includes('/content/home')) {
      setTimeout(() => {
        const targetElement = document.getElementById('services');
        if (targetElement) {
          targetElement.scrollIntoView({
            behavior: 'smooth',
            block: 'start',
            inline: 'nearest'
          });
          // Reset flag after successful scroll
          servicesBtnClickForScroll = false;
        }
      }, 200);
    }
  }
});

function setupPeriodicGlitches(glitchElement, terminalCursor) {
  // No more periodic glitch effects
}

function applyHeroEnhancements(heroStyle, glitchElement, terminalContent, terminalCursor, shouldType) {
  let calculatedDelay = 500; // Default delay
  
  if (heroStyle === 'cyberpunk') {
    // Apply both professional and cyberpunk effects
    applyProfessionalEffects(glitchElement, terminalContent, terminalCursor, false); // Don't duplicate boot sequence
    calculatedDelay = applyCyberpunkEffects(glitchElement, terminalContent, terminalCursor, shouldType);
  } else {
    // Professional mode only
    calculatedDelay = applyProfessionalEffects(glitchElement, terminalContent, terminalCursor, shouldType);
  }
  
  return calculatedDelay;
}

function applyProfessionalEffects(glitchElement, terminalContent, terminalCursor, shouldType) {
  // Professional effects - subtle and business-appropriate
  
  // 1. Subtle text glow during typing
  glitchElement.style.setProperty('--text-glow', '0 0 3px rgba(0, 255, 136, 0.3)');
  
  // 2. Enhanced cursor with subtle pulse
  terminalCursor.style.textShadow = '0 0 3px rgba(0, 255, 136, 0.5)';
  
  // 3. Add professional terminal class for styling
  glitchElement.classList.add('hero-professional');
  
  // Hide cursor immediately in professional mode
  terminalCursor.style.display = 'none';
  
  // 4. Boot sequence before quick fade-in (professional mode)
  debug("Professional effects - checking boot sequence conditions:", {
    textContent: terminalContent.textContent,
    shouldType: shouldType,
    initialAnimationsRun: initialAnimationsRun
  });
  
  if (shouldType && !initialAnimationsRun) {
    // Show the hero title immediately before boot sequence starts
    const fullText = "BLOCKHEAD CONSULTING";
    terminalContent.textContent = fullText;
    glitchElement.setAttribute('data-text', fullText);
    glitchElement.style.opacity = '1';
    terminalCursor.style.display = 'none';
    debug("Starting boot sequence in professional mode");
    // Start the boot sequence (it will overlay the hero title)
    const bootDuration = createBootSequence('professional', terminalContent);
    
    // Wait for boot sequence to completely fade out before starting other animations
    setTimeout(() => {
      // Fade in the subtitle after boot sequence has completely faded out
      const heroSubtitle = document.querySelector('.hero-subtitle');
      if (heroSubtitle) {
        heroSubtitle.classList.add('fade-in');
      }
    }, bootDuration + 100); // bootDuration already includes fade out time, add small buffer
    
    return bootDuration + 200; // Add small buffer
  }
  
  return 500; // Default delay for professional mode
}

function applyCyberpunkEffects(glitchElement, terminalContent, terminalCursor, shouldType) {
  // Cyberpunk effects - additional edgy elements while maintaining professionalism
  
  // 1. Enhanced glow effects
  glitchElement.style.setProperty('--text-glow', '0 0 8px rgba(0, 255, 136, 0.4)');
  
  // 2. Add cyberpunk class to hero section for CRT effects
  const heroSection = document.querySelector('.hero');
  if (heroSection) {
    heroSection.classList.add('hero-cyberpunk');
  }
  glitchElement.classList.add('hero-cyberpunk');
  
  // 3. Console message for cyberpunk mode
  debug('🌐 Cyberpunk Mode!');
  
  // 4. Boot sequence before typing (cyberpunk mode)
  if (terminalContent.textContent === '' && shouldType && !initialAnimationsRun) {
    const bootDuration = createBootSequence('cyberpunk', terminalContent);
    // Update the typing delay to account for boot sequence duration + fade out completion
    return bootDuration + 300; // Wait for boot sequence to completely fade out + small buffer
  }
  
  return 500; // Default delay for cyberpunk mode
}

function createMatrixRain() {
  // Create AI/crypto themed background instead of matrix rain
  createAICryptoBackground();
}

function createAICryptoBackground() {
  // Background elements removed - keeping only CRT effect
  // Mouse reactive animations disabled
}

// Background element functions moved to blockchain-animations.js (disabled)


// Old showBootSequence function removed - now using boot-sequence.js

// Form Encryption Animation
let formEncryptionRunning = false;
let encryptionIntervals = [];

function initializeFormEncryption() {
  // Find the contact form
  const contactForm = document.querySelector('.contact-form');
  if (!contactForm) {
    return;
  }
  
  // Remove any existing listeners first
  document.removeEventListener('htmx:beforeRequest', handleFormBeforeRequest);
  document.removeEventListener('htmx:afterRequest', handleFormAfterRequest);
  
  // Add document-level HTMX event listeners (recommended approach)
  document.addEventListener('htmx:beforeRequest', handleFormBeforeRequest);
  document.addEventListener('htmx:afterRequest', handleFormAfterRequest);
  
  // Add a global function for manual testing
  window.testFormEncryption = startFormEncryptionAnimation;
}

function handleFormBeforeRequest(e) {
  // Only handle contact form submissions
  if (!e.target.classList.contains('contact-form')) return;
  
  // Don't start animation if already running
  if (formEncryptionRunning) {
    e.preventDefault();
    return;
  }
  
  // Get form values for validation
  const form = e.target;
  const formData = new FormData(form);
  const values = {
    name: formData.get('name'),
    email: formData.get('email'), 
    message: formData.get('message')
  };
  
  // Client-side validation to prevent server errors
  const validationErrors = [];
  
  if (!values.name || values.name.trim().length < 2) {
    validationErrors.push('Name must be at least 2 characters long');
  }
  
  if (!values.email || !values.email.includes('@') || !values.email.includes('.')) {
    validationErrors.push('Please enter a valid email address');
  }
  
  if (!values.message || values.message.trim().length < 10) {
    validationErrors.push('Message must be at least 10 characters long');
  }
  
  if (validationErrors.length > 0) {
    e.preventDefault();
    
    // Show validation errors
    const responseDiv = document.getElementById('contact-response');
    if (responseDiv) {
      responseDiv.innerHTML = `<div class="alert error">${validationErrors.join('<br>')}</div>`;
    }
    return;
  }
}

function handleFormAfterRequest(e) {
  // Only handle contact form submissions
  if (!e.target.classList.contains('contact-form')) return;
  
  // Check if the response contains a success message (alternative way to detect success)
  const responseText = e.detail.xhr.responseText;
  const hasSuccessMessage = responseText && responseText.includes('Message sent successfully');
  
  // Only run animation on successful submission (check both ways)
  if ((e.detail.successful || hasSuccessMessage) && !formEncryptionRunning) {
    // Clear any HTMX success message first
    const responseDiv = document.getElementById('contact-response');
    if (responseDiv) {
      responseDiv.innerHTML = '';
    }
    startFormEncryptionAnimation();
  }
}

function startFormEncryptionAnimation() {
  if (formEncryptionRunning) return;
  
  // Clear any existing intervals first
  encryptionIntervals.forEach(interval => {
    clearInterval(interval);
  });
  encryptionIntervals = [];
  
  const contactForm = document.querySelector('.contact-form');
  const nameField = contactForm.querySelector('input[name="name"]');
  const emailField = contactForm.querySelector('input[name="email"]');
  const messageField = contactForm.querySelector('textarea[name="message"]');
  const submitButton = contactForm.querySelector('button[type="submit"]');
  const responseDiv = document.getElementById('contact-response');
  
  if (!nameField || !emailField || !messageField) return;
  
  formEncryptionRunning = true;
  
  // Store original values
  const originalValues = {
    name: nameField.value,
    email: emailField.value,
    message: messageField.value
  };
  
  // Add encrypting class to form
  contactForm.classList.add('form-encrypting');
  
  // Hide any existing response
  if (responseDiv) {
    responseDiv.innerHTML = '';
  }
  
  // Start the encryption cycle
  let cycleCount = 0;
  const maxCycles = 3;
  
  function runEncryptionCycle() {
    cycleCount++;
    
    // Phase 1: Scramble text (500ms)
    scrambleFields([nameField, emailField, messageField], originalValues);
    
    setTimeout(() => {
      // Phase 2: Glitch effect (400ms)
      glitchFields([nameField, emailField, messageField]);
      
      setTimeout(() => {
        // Phase 3: Show "Encrypting..." text (700ms)
        showEncryptingText([nameField, emailField, messageField]);
        
        setTimeout(() => {
          if (cycleCount < maxCycles) {
            // Continue to next cycle
            runEncryptionCycle();
          } else {
            // Final cycle complete - clear fields and show success
            finishEncryption([nameField, emailField, messageField], responseDiv, contactForm);
          }
        }, 700);
        
      }, 400);
      
    }, 500);
  }
  
  // Start the first cycle
  runEncryptionCycle();
}

function scrambleFields(fields, originalValues) {
  fields.forEach(field => {
    field.classList.add('scrambling');
    field.classList.remove('glitching', 'encrypting-text');
    
    const originalValue = originalValues[field.name];
    let scrambledText = '';
    
    // Create scrambled version of the text
    for (let i = 0; i < originalValue.length; i++) {
      const char = originalValue[i];
      if (char === ' ') {
        scrambledText += ' ';
      } else {
        // Random character substitution
        const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()';
        scrambledText += chars[Math.floor(Math.random() * chars.length)];
      }
    }
    
    field.value = scrambledText;
  });
}

function glitchFields(fields) {
  fields.forEach(field => {
    field.classList.add('glitching');
    field.classList.remove('scrambling', 'encrypting-text');
    
    // Add more random corruption
    let glitchedText = field.value;
    const corruptionChars = '█▓▒░▄▀■□▪▫';
    
    // Replace some characters with corruption symbols
    glitchedText = glitchedText.split('').map(char => {
      if (Math.random() < 0.3) {
        return corruptionChars[Math.floor(Math.random() * corruptionChars.length)];
      }
      return char;
    }).join('');
    
    field.value = glitchedText;
  });
}

function showEncryptingText(fields) {
  fields.forEach(field => {
    field.classList.add('encrypting-text');
    field.classList.remove('scrambling', 'glitching');
    
    // Show encrypting status with animated dots
    const messages = ['Encrypting.', 'Encrypting..', 'Encrypting...'];
    let dotCount = 0;
    
    const interval = setInterval(() => {
      field.value = messages[dotCount % messages.length];
      dotCount++;
      
      if (dotCount >= messages.length * 2) {
        clearInterval(interval);
        // Remove from intervals array when done
        const index = encryptionIntervals.indexOf(interval);
        if (index > -1) {
          encryptionIntervals.splice(index, 1);
        }
      }
    }, 120);
    
    // Store interval so it can be cleared if needed
    encryptionIntervals.push(interval);
  });
}

function finishEncryption(fields, responseDiv, contactForm) {
  // Clear any running intervals first
  encryptionIntervals.forEach(interval => {
    clearInterval(interval);
  });
  encryptionIntervals = [];
  
  // Clear all fields completely
  fields.forEach(field => {
    field.value = '';
    field.classList.remove('scrambling', 'glitching', 'encrypting-text');
  });
  
  // Remove encrypting class from form (this should re-enable button via CSS)
  contactForm.classList.remove('form-encrypting');
  
  // Show success message with fade-out
  if (responseDiv) {
    responseDiv.innerHTML = `
      <div class="alert success" id="encryption-success-message">
        <strong>Encrypted message sent successfully!</strong><br>
        I'll get back to you within 24 hours.
      </div>
    `;
    
    // Auto-fade success message after 4 seconds
    setTimeout(() => {
      const successMessage = document.getElementById('encryption-success-message');
      if (successMessage) {
        successMessage.style.transition = 'opacity 1s ease-out';
        successMessage.style.opacity = '0';
        
        // Remove from DOM after fade completes
        setTimeout(() => {
          if (responseDiv && successMessage.parentNode === responseDiv) {
            responseDiv.innerHTML = '';
          }
        }, 1000);
      }
    }, 4000);
  }
  
  // Reset flag - form is ready for new submissions
  formEncryptionRunning = false;
}


