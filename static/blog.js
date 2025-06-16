// Blog-specific functionality

// Blog filters and search
function initializeBlogFilters() {
  const tagButtons = document.querySelectorAll('.filter-btn');
  const blogPosts = document.querySelectorAll('.blog-post-card');
  const searchInput = document.getElementById('blog-search');
  const searchClear = document.querySelector('.search-clear');
  
  // Tag filtering
  tagButtons.forEach(button => {
    button.addEventListener('click', function() {
      const tag = this.dataset.tag.toLowerCase();
      
      // Update active state
      tagButtons.forEach(btn => btn.classList.remove('active'));
      this.classList.add('active');
      
      // Use blog filter configuration from server (injected by template)
      const tagAliases = window.blogFilterConfig || {};
      
      blogPosts.forEach(post => {
        const postTags = post.dataset.tags.toLowerCase();
        
        if (tag === 'all') {
          post.style.display = 'block';
          return;
        }
        
        // Check if any alias matches (using word boundaries to avoid partial matches)
        const aliases = tagAliases[tag.toLowerCase()] || [tag.toLowerCase()];
        const matches = aliases.some(alias => {
          const regex = new RegExp(`\\b${alias.toLowerCase().replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\b`);
          return regex.test(postTags);
        });
        
        
        if (matches) {
          post.style.display = 'block';
        } else {
          post.style.display = 'none';
        }
      });
      
      // Update URL without page reload
      const url = new URL(window.location);
      if (tag === 'all') {
        url.searchParams.delete('tag');
      } else {
        url.searchParams.set('tag', tag);
      }
      window.history.pushState({}, '', url);
    });
  });
  
  // Search functionality
  if (searchInput) {
    searchInput.addEventListener('input', function() {
      const searchTerm = this.value.toLowerCase();
      
      // Show/hide clear button
      if (searchClear) {
        searchClear.style.display = searchTerm ? 'block' : 'none';
      }
      
      // Filter posts
      blogPosts.forEach(post => {
        const title = post.querySelector('.blog-title').textContent.toLowerCase();
        const summary = post.querySelector('.blog-summary').textContent.toLowerCase();
        const tags = post.dataset.tags.toLowerCase();
        
        if (title.includes(searchTerm) || summary.includes(searchTerm) || tags.includes(searchTerm)) {
          post.style.display = 'block';
        } else {
          post.style.display = 'none';
        }
      });
    });
    
    // Clear search
    if (searchClear) {
      searchClear.addEventListener('click', function() {
        searchInput.value = '';
        searchInput.dispatchEvent(new Event('input'));
      });
    }
  }
  
  // Check URL for initial tag filter
  const urlParams = new URLSearchParams(window.location.search);
  const tagParam = urlParams.get('tag');
  if (tagParam) {
    const tagButton = document.querySelector(`[data-tag="${tagParam}"]`);
    if (tagButton) {
      tagButton.click();
    }
  }
}

// Project Card System
function initializeProjectCards() {
  // Look for tables with the special class
  const projectTables = document.querySelectorAll('.blog-content table.project-table');
  
  projectTables.forEach(table => {
    // Find all project detail sections that follow this table
    let nextElement = table.nextElementSibling;
    const projectDetails = {};
    
    // Skip any h3/h4 headers that might be between table and details
    while (nextElement && (nextElement.tagName === 'H3' || nextElement.tagName === 'H4')) {
      nextElement = nextElement.nextElementSibling;
    }
    
    // Collect all project details until we hit something that's not a project detail
    while (nextElement && 
           ((nextElement.tagName === 'P' && nextElement.querySelector('strong')) || 
           nextElement.tagName === 'P' || 
           nextElement.tagName === 'UL')) {
      
      // Check if this is a project header (bold text)
      if (nextElement.tagName === 'P' && nextElement.querySelector('strong')) {
        const strongElement = nextElement.querySelector('strong');
        const projectName = strongElement.textContent.split('(')[0].trim();
        
        // Initialize project data
        projectDetails[projectName] = {
          header: nextElement,
          metadata: [],
          content: []
        };
        
        // Collect following elements until next project or end
        let detailElement = nextElement.nextElementSibling;
        while (detailElement && 
               !(detailElement.tagName === 'P' && detailElement.querySelector('strong'))) {
          
          // Check if it's metadata (italic text)
          if (detailElement.tagName === 'P' && detailElement.querySelector('em')) {
            projectDetails[projectName].metadata.push(detailElement);
          } else if (detailElement.tagName === 'P' || detailElement.tagName === 'UL') {
            projectDetails[projectName].content.push(detailElement);
          } else {
            break; // Stop if we hit something unexpected
          }
          
          detailElement = detailElement.nextElementSibling;
        }
      }
      
      nextElement = nextElement.nextElementSibling;
    }
    
    // Hide all project details initially
    Object.values(projectDetails).forEach(project => {
      project.header.style.display = 'none';
      project.metadata.forEach(el => el.style.display = 'none');
      project.content.forEach(el => el.style.display = 'none');
    });
    
    // Create card container
    const cardContainer = document.createElement('div');
    cardContainer.className = 'project-card-container';
    cardContainer.style.display = 'none';
    table.parentNode.insertBefore(cardContainer, table.nextElementSibling);
    
    // Add click handlers to table rows
    const rows = table.querySelectorAll('tbody tr');
    rows.forEach(row => {
      row.style.cursor = 'pointer';
      
      row.addEventListener('click', () => {
        const projectName = row.cells[0].textContent.trim();
        const projectData = projectDetails[projectName];
        
        if (projectData) {
          // Clear existing content
          cardContainer.innerHTML = '';
          
          // Create card
          const card = document.createElement('div');
          card.className = 'project-detail-card';
          
          // Add close button
          const closeBtn = document.createElement('button');
          closeBtn.className = 'project-card-close';
          closeBtn.innerHTML = '×';
          closeBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            cardContainer.style.display = 'none';
            // Remove active state from all rows
            rows.forEach(r => r.classList.remove('active'));
          });
          card.appendChild(closeBtn);
          
          // Clone and add content
          const headerClone = projectData.header.cloneNode(true);
          headerClone.style.display = 'block';
          card.appendChild(headerClone);
          
          projectData.metadata.forEach(el => {
            const clone = el.cloneNode(true);
            clone.style.display = 'block';
            card.appendChild(clone);
          });
          
          projectData.content.forEach(el => {
            const clone = el.cloneNode(true);
            clone.style.display = 'block';
            card.appendChild(clone);
          });
          
          cardContainer.appendChild(card);
          cardContainer.style.display = 'block';
          
          // Add active state to clicked row
          rows.forEach(r => r.classList.remove('active'));
          row.classList.add('active');
          
          // Smooth scroll to card
          setTimeout(() => {
            cardContainer.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
          }, 100);
        }
      });
    });
  });
}

// Initialize blog features
function initializeBlog() {
  initializeBlogFilters();
  initializeProjectCards();
}

// Auto-initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
  initializeBlog();
});

// Re-initialize on HTMX navigation
document.addEventListener('htmx:afterSwap', (evt) => {
  // Only initialize for blog-related pages
  if (evt.detail.target.id === 'main-content' && 
      (evt.detail.xhr.responseURL.includes('/blog') || 
       evt.detail.xhr.responseURL.includes('/content/blog'))) {
    setTimeout(() => {
      initializeBlog();
    }, 50);
  }
});

// Export for testing
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    initializeBlog,
    initializeBlogFilters,
    initializeProjectCards
  };
}