// Boot Sequence Configuration and Logic
// This file contains configurable boot sequences for different hero styles

// Conditional console logging - will be set from main.js
function debugBoot(...args) {
  if (typeof debugLogging !== 'undefined' && debugLogging) {
    console.log(...args);
  }
}

function debugBootError(...args) {
  if (typeof debugLogging !== 'undefined' && debugLogging) {
    console.error(...args);
  }
}

// Boot sequences will be injected from server config
const BOOT_SEQUENCES = window.bootSequences || {
	professional: [
		"Initializing secure blockchain infrastructure.",
		"Loading enterprise AI integration systems.",
		"Connecting to production-grade payment rails.",
		"Strategic consulting protocols online.",
		"Ready for high-impact collaboration.",
	],
	
	professionalMobile: [
		"Integrating AI",
		"Analyzing Chain Data",
		"Maximizing Engineering Spend",
		"Ready to help you win",
	],

	cyberpunk: [
		"System initializing...",
		"Loading neural networks...",
		"Establishing secure connection...",
		"Ready.",
	],
	
	cyberpunkMobile: [
		"Booting...",
		"Neural link active",
		"Chain sync complete",
		"Ready.",
	],
};

const BOOT_CONFIG = {
	professional: {
		messages: BOOT_SEQUENCES.professional,
		messageDelay: 1200, // 1.2 seconds between messages (more readable pace for longer messages)
		finalPause: 1000, // Longer pause before fade-out
		fadeOutDuration: 800, // Fade-out animation duration
		mobileMessages: BOOT_SEQUENCES.professionalMobile,
		mobileFontSize: '1.2rem', // Smaller font for mobile
	},

	cyberpunk: {
		messages: BOOT_SEQUENCES.cyberpunk,
		messageDelay: 800, // 0.8 seconds between messages
		finalPause: 600, // Pause before fade-out
		fadeOutDuration: 800, // Fade-out animation duration
		mobileMessages: BOOT_SEQUENCES.cyberpunkMobile,
		mobileFontSize: '1.2rem', // Smaller font for mobile
	},
};

function createBootSequence(heroStyle = "professional", terminalContent) {
	debugBoot("Boot sequence starting for", heroStyle, "at", new Date().toLocaleTimeString());
	const config = BOOT_CONFIG[heroStyle] || BOOT_CONFIG.professional;

	// Check if boot sequence is already running
	if (document.querySelector('.boot-sequence')) {
		debugBoot("Boot sequence already exists, skipping");
		return 0;
	}

	// Find the hero title element to position the boot sequence in the same location
	const glitchElement = document.querySelector(".glitch");
	if (!glitchElement) {
		debugBootError("Boot sequence: glitch element not found");
		return 0;
	}

	// Detect mobile viewport
	const isMobile = window.innerWidth <= 768;
	const messages = isMobile && config.mobileMessages ? config.mobileMessages : config.messages;
	const fontSize = isMobile && config.mobileFontSize ? config.mobileFontSize : '1.5rem';

	// Find the longest message to calculate optimal width
	const longestMessage = messages.reduce(
		(longest, current) => (current.length > longest.length ? current : longest),
		"",
	);

	// Create a temporary element to measure text width
	const tempElement = document.createElement("div");
	tempElement.style.cssText = `
    position: absolute;
    visibility: hidden;
    white-space: nowrap;
    font-family: var(--font-mono);
    font-size: ${fontSize};
    font-weight: 400;
    letter-spacing: ${isMobile ? '0.01em' : '0.02em'};
  `;

	tempElement.textContent = longestMessage + "..."; // Add dots for accurate measurement
	document.body.appendChild(tempElement);
	const textWidth = tempElement.offsetWidth;
	document.body.removeChild(tempElement);

	// Create a separate boot sequence element
	const bootElement = document.createElement("div");
	bootElement.className = "terminal-boot-sequence boot-sequence";
	// Find the subtitle element to position boot sequence before it
	const heroSubtitle = document.querySelector('.hero-subtitle');
	
	bootElement.style.cssText = `
    position: absolute;
    width: 100%;
    left: 0;
    top: 100%;
    margin-top: 0.5rem;
    font-family: var(--font-mono);
    font-size: ${fontSize};
    font-weight: 400;
    color: #888888;
    text-align: center;
    line-height: 1.4;
    text-shadow: 0 0 10px rgba(136, 136, 136, 0.3);
    letter-spacing: ${isMobile ? '0.01em' : '0.02em'};
    white-space: nowrap;
    opacity: 1;
    padding: 0 1rem;
    z-index: 10;
  `;

	// No additional mobile styles needed - already handled above

	// Make sure the glitch element is positioned for absolute children
	glitchElement.style.position = 'relative';
	
	// Insert the boot element as a child of hero title so it's positioned relative to it
	debugBoot("Inserting boot element as child of hero title");
	glitchElement.appendChild(bootElement);
	debugBoot("Boot element added to DOM:", bootElement);

	let currentMessage = 0;
	let dotCount = 0;

	function showNextMessage() {
		if (currentMessage < messages.length) {
			const message = messages[currentMessage];
			bootElement.textContent = message;
			dotCount = 0;

			// Add sequential dots effect
			function addDots() {
				if (dotCount < 3) {
					dotCount++;
					bootElement.textContent = message + ".".repeat(dotCount);
					setTimeout(addDots, 200);
				} else {
					// Move to next message after dots
					currentMessage++;
					if (currentMessage < messages.length) {
						setTimeout(showNextMessage, config.messageDelay - 600); // Account for dot animation time
					} else {
						// Fade out the boot sequence after final pause
						setTimeout(() => {
							debugBoot("Starting fade out of boot sequence");
							bootElement.style.transition = `opacity ${config.fadeOutDuration}ms ease-out`;
							bootElement.style.opacity = "0";
							setTimeout(() => {
								debugBoot("Removing boot element from DOM");
								if (bootElement && bootElement.parentElement) {
									bootElement.parentElement.removeChild(bootElement);
									debugBoot("Boot element removed successfully");
								} else {
									debugBootError("Boot element has no parent - cannot remove");
								}
							}, config.fadeOutDuration);
						}, config.finalPause);
					}
				}
			}

			// Start adding dots after showing the message
			setTimeout(addDots, 300);
		}
	}

	// Start boot sequence after a brief delay
	setTimeout(showNextMessage, 300);

	// Return total duration for timing other animations
	const totalDuration =
		300 +
		messages.length * config.messageDelay +
		config.finalPause +
		config.fadeOutDuration;

	return totalDuration;
}

// Export for potential module use
if (typeof module !== "undefined" && module.exports) {
	module.exports = {
		createBootSequence,
		BOOT_SEQUENCES,
		BOOT_CONFIG,
	};
}

