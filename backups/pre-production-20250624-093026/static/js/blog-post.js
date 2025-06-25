// Blog Post Functionality
// Handles Mermaid initialization and post navigation

// Prevent redeclaration if script is loaded multiple times
if (typeof window.BlogPost === 'undefined') {
window.BlogPost = {
  initializeMermaid() {
    // Check if mermaid is available
    if (typeof mermaid === 'undefined') {
      console.warn('Mermaid library not loaded');
      return;
    }

    // Check if there are any client-side diagrams
    const csrDiagrams = document.querySelectorAll('.mermaid-csr .mermaid');
    
    if (csrDiagrams.length === 0) {
      return;
    }

    // Initialize Mermaid with the exact original configuration
    mermaid.initialize({
      startOnLoad: false,  // We'll manually render to avoid conflicts
      theme: 'default',
      themeVariables: {
        // Dark theme exactly matching code block styling
        primaryColor: '#a6e22e',           // Bright green like code syntax
        primaryTextColor: '#f8f8f2',       // Light gray like code text
        primaryBorderColor: '#66d9ef',     // Blue like code keywords
        lineColor: '#f92672',              // Pink/red like code operators
        secondaryColor: '#66d9ef',         // Blue for secondary elements
        tertiaryColor: '#272822',          // Dark background like code blocks
        background: '#272822',             // Dark background like code blocks
        mainBkg: '#3c3d38',                // Slightly lighter for nodes (like line highlight)
        secondBkg: '#3c3d38',              // Same as mainBkg
        tertiaryBkg: '#272822',            // Dark background
        
        // Light text colors matching code syntax
        textColor: '#f8f8f2',              // Light gray like code text
        nodeTextColor: '#f8f8f2',          // Light gray text in nodes
        labelTextColor: '#f8f8f2',         // Light gray labels
        
        // Accent colors from syntax highlighting
        loopTextColor: '#a6e22e',          // Green for loop labels
        activationBorderColor: '#f92672',  // Pink/red borders
        altColor: '#3c3d38',               // Slightly lighter background for alt sections
        
        // Additional colors for various diagram types
        actorBkg: '#3c3d38',               // Slightly lighter for actors
        actorBorder: '#66d9ef',            // Blue borders for actors
        noteBkg: '#3c3d38',                // Note background
        noteBorder: '#f92672',             // Pink border for notes
        noteTextColor: '#f8f8f2',          // Light text in notes
        sectionBkgColor: '#272822',        // Section background
        cScale0: '#f92672',                // Pink
        cScale1: '#a6e22e',                // Green
        cScale2: '#66d9ef',                // Blue
        cScale3: '#fd971f',                // Orange
        cScale4: '#ae81ff',                // Purple
        cScale5: '#e6db74',                // Yellow
        pie1: '#f92672',                   // Pink slice
        pie2: '#a6e22e',                   // Green slice
        pie3: '#66d9ef',                   // Blue slice
        pie4: '#fd971f',                   // Orange slice
        pie5: '#ae81ff',                   // Purple slice
        pie6: '#e6db74',                   // Yellow slice
        pie7: '#f8f8f2',                   // Light gray slice
        pieTitleTextSize: 18,              // Larger title
        pieTitleTextColor: '#f8f8f2',      // Light title
        pieLegendTextSize: 14,             // Larger legend
        pieLegendTextColor: '#f8f8f2',     // Light legend
        pieStrokeColor: '#272822',         // Dark stroke
        pieStrokeWidth: 2,                 // Thicker stroke
        pieOuterStrokeWidth: 2,            // Outer stroke
        
        // State diagram colors
        errorBkgColor: '#f92672',          // Pink error background
        errorTextColor: '#f8f8f2',         // Light error text
        
        // Sequence diagram
        actorLineColor: '#66d9ef',         // Blue actor line
        signalColor: '#f8f8f2',            // Light signal color
        signalTextColor: '#f8f8f2',        // Light signal text
        labelBoxBkgColor: '#3c3d38',       // Label box background
        labelBoxBorderColor: '#66d9ef',    // Label box border
        labelTextColor: '#f8f8f2',         // Label text
        activationBkgColor: '#3c3d38',     // Activation background
        
        // Flowchart
        edgeLabelBackground: '#272822',    // Edge label background
        
        // ER diagram
        relationColor: '#f92672',          // Relation color
        
        // Class diagram
        classText: '#f8f8f2',              // Class text
        
        // Git diagram
        commitLabelFontSize: 10,           // Smaller commit labels
        git0: '#f92672',                   // Branch 0 color
        git1: '#a6e22e',                   // Branch 1 color
        git2: '#66d9ef',                   // Branch 2 color
        git3: '#fd971f',                   // Branch 3 color
        git4: '#ae81ff',                   // Branch 4 color
        git5: '#e6db74',                   // Branch 5 color
        git6: '#f8f8f2',                   // Branch 6 color
        git7: '#75715e',                   // Branch 7 color
        gitBranchLabel0: '#f8f8f2',        // Branch label color
        gitBranchLabel1: '#f8f8f2',        
        gitBranchLabel2: '#f8f8f2',        
        gitBranchLabel3: '#f8f8f2',        
        gitBranchLabel4: '#f8f8f2',        
        gitBranchLabel5: '#f8f8f2',        
        gitBranchLabel6: '#f8f8f2',        
        gitBranchLabel7: '#f8f8f2',        
        
        // Requirement diagram
        requirementBackground: '#3c3d38',  // Requirement background
        requirementBorderColor: '#66d9ef', // Requirement border
        requirementBorderSize: 1,          // Border size
        requirementTextColor: '#f8f8f2',   // Requirement text
        relationLabelBackground: '#272822', // Relation label background
        relationLabelColor: '#f8f8f2'      // Relation label color
      },
      flowchart: {
        htmlLabels: true,
        curve: 'basis',
        padding: 15,                     // More padding for better visibility
        useMaxWidth: true,
        rankSpacing: 50,                 // More space between ranks
        nodeSpacing: 30,                 // More space between nodes
        fontSize: 12                     // Slightly larger font
      },
      sequence: {
        diagramMarginX: 50,
        diagramMarginY: 30,
        actorMargin: 50,
        width: 200,
        height: 60,
        boxMargin: 15,
        boxTextMargin: 8,
        noteMargin: 15,
        messageMargin: 40,
        mirrorActors: true,
        bottomMarginAdj: 1,
        useMaxWidth: true,
        rightAngles: false,
        showSequenceNumbers: false,
        actorFontSize: 12,               // Larger actor font
        actorFontFamily: 'JetBrains Mono, Fira Code, monospace',
        noteFontSize: 12,                // Larger note font
        messageFontSize: 12              // Larger message font
      },
      journey: {
        diagramMarginX: 50,
        diagramMarginY: 30,
        leftMargin: 150,
        width: 200,
        height: 60,
        boxMargin: 15,
        boxTextMargin: 8,
        noteMargin: 15,
        messageMargin: 40,
        bottomMarginAdj: 1,
        useMaxWidth: true,
        actorFontSize: 12,               // Larger actor font
        sectionFontSize: 14,             // Larger section font
        noteFontSize: 12                 // Larger note font
      },
      gantt: {
        leftPadding: 85,
        gridLineStartPadding: 40,
        fontSize: 12,                    // Slightly larger font
        fontFamily: 'JetBrains Mono, Fira Code, monospace'
      },
      // General settings
      deterministicIds: true
    });
    
    // Manually render each client-side diagram immediately (no delays)
    csrDiagrams.forEach(async (element, index) => {
      // Skip if already rendered
      if (element.querySelector('svg') || element.classList.contains('mermaid-processed')) {
        return;
      }
      
      try {
        const graphDefinition = element.textContent.trim();
        
        // Generate a unique ID for this render
        const graphId = `mermaid-csr-${Date.now()}-${index}`;
        
        // Mark as processed to prevent double rendering
        element.classList.add('mermaid-processed');
        
        // Render the diagram immediately
        const { svg } = await mermaid.render(graphId, graphDefinition);
        
        // Replace the original element content with the rendered SVG
        element.innerHTML = svg;
      } catch (error) {
        console.error('Mermaid rendering error for diagram', index, ':', error);
        element.innerHTML = '<div class="mermaid-error">Error rendering diagram: ' + error.message + '</div>';
      }
    });
  },

  loadPostNavigation() {
    const currentSlug = window.location.pathname.split('/').pop();
    
    // Fetch and display post navigation
    fetch(`/blog/${currentSlug}/navigation`)
      .then(response => response.json())
      .then(data => {
        const prevDiv = document.querySelector('.post-nav-prev');
        const nextDiv = document.querySelector('.post-nav-next');
        
        if (!prevDiv || !nextDiv) return;
        
        if (data.Previous) {
          prevDiv.innerHTML = `
            <a href="/blog/${data.Previous.slug}" class="post-nav-link prev" title="${data.Previous.title}">
              <span class="nav-direction">← Previous</span>
            </a>
          `;
        } else {
          prevDiv.innerHTML = '<div class="post-nav-placeholder"></div>';
        }
        
        if (data.Next) {
          nextDiv.innerHTML = `
            <a href="/blog/${data.Next.slug}" class="post-nav-link next" title="${data.Next.title}">
              <span class="nav-direction">Next →</span>
            </a>
          `;
        } else {
          nextDiv.innerHTML = '<div class="post-nav-placeholder"></div>';
        }
      })
      .catch(error => {
        console.error('Error loading post navigation:', error);
      });
  },

  loadRelatedPosts() {
    const currentSlug = window.location.pathname.split('/').pop();
    
    // Fetch and display related posts
    fetch(`/blog/${currentSlug}/related?limit=2`)
      .then(response => response.json())
      .then(data => {
        const relatedContainer = document.querySelector('.related-posts');
        if (!relatedContainer || !data || data.length === 0) return;
        
        const relatedHTML = data.map(item => {
          const date = new Date(item.Post.date);
          const formattedDate = date.toLocaleDateString('en-US', { 
            year: 'numeric', 
            month: 'short', 
            day: 'numeric' 
          });
          
          return `
            <div class="related-post-card">
              <a href="/blog/${item.Post.slug}" class="related-post-link">
                <h4>${item.Post.title}</h4>
                <p>${item.Post.summary}</p>
                <div class="post-meta">
                  <span class="post-date">${formattedDate}</span>
                  <span class="post-reading-time">${item.Post.reading_time || 5} min read</span>
                </div>
              </a>
            </div>
          `;
        }).join('');
        
        relatedContainer.innerHTML = `
          <h3>Related Posts</h3>
          <div class="related-posts-grid">
            ${relatedHTML}
          </div>
        `;
      })
      .catch(error => {
        console.error('Error loading related posts:', error);
      });
  },

  initialize() {
    this.initializeMermaid();
    this.loadPostNavigation();
    this.loadRelatedPosts();
  }
};
}

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
  if (document.querySelector('.blog-post')) {
    window.BlogPost.initialize();
  }
});

// Re-initialize after HTMX content swaps
document.addEventListener('htmx:afterSwap', (event) => {
  if (event.detail.target.querySelector && event.detail.target.querySelector('.blog-post')) {
    window.BlogPost.initialize();
  }
});

// Also handle HTMX before swaps to ensure Mermaid is available
document.addEventListener('htmx:beforeSwap', (event) => {
  // Preload Mermaid if we're navigating to a blog post
  const response = event.detail.xhr.responseText;
  if (response && response.includes('blog-post') && response.includes('mermaid')) {
    if (typeof mermaid === 'undefined' && !document.querySelector('script[src*="mermaid"]')) {
      const script = document.createElement('script');
      script.src = 'https://unpkg.com/mermaid@11/dist/mermaid.min.js';
      document.head.appendChild(script);
    }
  }
});