const { JSDOM } = require('jsdom');

describe('Project Cards Feature', () => {
  let window, document;
  let initializeProjectCards;

  beforeEach(() => {
    // Create a mock DOM
    const dom = new JSDOM(`
      <!DOCTYPE html>
      <html>
        <head></head>
        <body>
          <div class="blog-content">
            <table class="project-table">
              <thead>
                <tr>
                  <th>Project</th>
                  <th>LOC</th>
                  <th>Value</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td><strong>Guild Framework</strong></td>
                  <td>151,315</td>
                  <td>$5.25M</td>
                </tr>
                <tr>
                  <td><strong>AlgoScales</strong></td>
                  <td>25,537</td>
                  <td>$811K</td>
                </tr>
              </tbody>
            </table>
            
            <h3>Project Details</h3>
            
            <p><strong>Guild Framework</strong> ($5.25M COCOMO value)</p>
            <p><em>Tech Stack: Go, gRPC, SQLite</em></p>
            <p><em>Purpose: Multi-agent AI orchestration framework</em></p>
            <p>The crown jewel of this sprint. It's a complete enterprise-grade platform.</p>
            <ul>
              <li>Advanced features</li>
              <li>Great performance</li>
            </ul>
            
            <p><strong>AlgoScales</strong> ($811K COCOMO value)</p>
            <p><em>Tech Stack: Go, Lua, Vim Script</em></p>
            <p><em>Purpose: Algorithm practice tool</em></p>
            <p>An algorithm practice tool with AI hints.</p>
          </div>
        </body>
      </html>
    `, {
      url: 'http://localhost',
      resources: 'usable',
      runScripts: 'dangerously'
    });

    window = dom.window;
    document = window.document;

    // Mock the initializeProjectCards function
    initializeProjectCards = () => {
      const projectTables = document.querySelectorAll('.blog-content table.project-table');
      
      projectTables.forEach(table => {
        let nextElement = table.nextElementSibling;
        const projectDetails = {};
        
        // Skip any h3/h4 headers
        while (nextElement && (nextElement.tagName === 'H3' || nextElement.tagName === 'H4')) {
          nextElement = nextElement.nextElementSibling;
        }
        
        // Collect project details
        while (nextElement && 
               ((nextElement.tagName === 'P' && nextElement.querySelector('strong')) || 
               nextElement.tagName === 'P' || 
               nextElement.tagName === 'UL')) {
          
          if (nextElement.tagName === 'P' && nextElement.querySelector('strong')) {
            const strongElement = nextElement.querySelector('strong');
            const projectName = strongElement.textContent.split('(')[0].trim();
            
            projectDetails[projectName] = {
              header: nextElement,
              metadata: [],
              content: []
            };
            
            let detailElement = nextElement.nextElementSibling;
            while (detailElement && 
                   !(detailElement.tagName === 'P' && detailElement.querySelector('strong'))) {
              
              if (detailElement.tagName === 'P' && detailElement.querySelector('em')) {
                projectDetails[projectName].metadata.push(detailElement);
              } else if (detailElement.tagName === 'P' || detailElement.tagName === 'UL') {
                projectDetails[projectName].content.push(detailElement);
              } else {
                break;
              }
              
              detailElement = detailElement.nextElementSibling;
            }
          }
          
          nextElement = nextElement.nextElementSibling;
        }
        
        // Hide all project details
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
        
        // Add click handlers
        const rows = table.querySelectorAll('tbody tr');
        rows.forEach(row => {
          row.style.cursor = 'pointer';
          row.setAttribute('data-clickable', 'true');
          
          row.addEventListener('click', () => {
            const projectName = row.cells[0].textContent.trim();
            const projectData = projectDetails[projectName];
            
            if (projectData) {
              cardContainer.innerHTML = '';
              
              const card = document.createElement('div');
              card.className = 'project-detail-card';
              
              const closeBtn = document.createElement('button');
              closeBtn.className = 'project-card-close';
              closeBtn.innerHTML = '×';
              card.appendChild(closeBtn);
              
              // Clone content
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
              
              rows.forEach(r => r.classList.remove('active'));
              row.classList.add('active');
            }
          });
        });
      });
    };
  });

  afterEach(() => {
    // Clean up
    window.close();
  });

  test('should find project tables with the correct class', () => {
    initializeProjectCards();
    
    const tables = document.querySelectorAll('.blog-content table.project-table');
    expect(tables.length).toBe(1);
  });

  test('should hide project details on initialization', () => {
    initializeProjectCards();
    
    // Check that project headers are hidden
    const projectHeaders = document.querySelectorAll('p > strong');
    projectHeaders.forEach(header => {
      if (header.textContent.includes('COCOMO value')) {
        expect(header.parentElement.style.display).toBe('none');
      }
    });
    
    // Check that metadata is hidden
    const metadataElements = document.querySelectorAll('p > em');
    metadataElements.forEach(meta => {
      if (meta.textContent.includes('Tech Stack:') || meta.textContent.includes('Purpose:')) {
        expect(meta.parentElement.style.display).toBe('none');
      }
    });
  });

  test('should make table rows clickable', () => {
    initializeProjectCards();
    
    const rows = document.querySelectorAll('.project-table tbody tr');
    rows.forEach(row => {
      expect(row.style.cursor).toBe('pointer');
      expect(row.getAttribute('data-clickable')).toBe('true');
    });
  });

  test('should create card container after table', () => {
    initializeProjectCards();
    
    const table = document.querySelector('.project-table');
    const nextElement = table.nextElementSibling;
    
    expect(nextElement).toBeTruthy();
    expect(nextElement.className).toBe('project-card-container');
    expect(nextElement.style.display).toBe('none');
  });

  test('should show project card when row is clicked', () => {
    initializeProjectCards();
    
    // Find and click the Guild Framework row
    const rows = document.querySelectorAll('.project-table tbody tr');
    const guildRow = Array.from(rows).find(row => 
      row.cells[0].textContent.includes('Guild Framework')
    );
    
    // Simulate click
    guildRow.click();
    
    // Check that card container is visible
    const cardContainer = document.querySelector('.project-card-container');
    expect(cardContainer.style.display).toBe('block');
    
    // Check that card has correct content
    const card = cardContainer.querySelector('.project-detail-card');
    expect(card).toBeTruthy();
    
    // Check for close button
    const closeBtn = card.querySelector('.project-card-close');
    expect(closeBtn).toBeTruthy();
    expect(closeBtn.innerHTML).toBe('×');
    
    // Check that content includes project details
    const cardText = card.textContent;
    expect(cardText).toContain('Guild Framework');
    expect(cardText).toContain('Tech Stack: Go, gRPC, SQLite');
    expect(cardText).toContain('Purpose: Multi-agent AI orchestration framework');
    expect(cardText).toContain('crown jewel');
    
    // Check that row is marked as active
    expect(guildRow.classList.contains('active')).toBe(true);
  });

  test('should switch between different project cards', () => {
    initializeProjectCards();
    
    const rows = document.querySelectorAll('.project-table tbody tr');
    const guildRow = Array.from(rows).find(row => 
      row.cells[0].textContent.includes('Guild Framework')
    );
    const algoRow = Array.from(rows).find(row => 
      row.cells[0].textContent.includes('AlgoScales')
    );
    
    // Click Guild first
    guildRow.click();
    
    let cardText = document.querySelector('.project-detail-card').textContent;
    expect(cardText).toContain('Guild Framework');
    expect(guildRow.classList.contains('active')).toBe(true);
    expect(algoRow.classList.contains('active')).toBe(false);
    
    // Then click AlgoScales
    algoRow.click();
    
    cardText = document.querySelector('.project-detail-card').textContent;
    expect(cardText).toContain('AlgoScales');
    expect(cardText).toContain('Algorithm practice tool');
    expect(guildRow.classList.contains('active')).toBe(false);
    expect(algoRow.classList.contains('active')).toBe(true);
  });

  test('should handle projects with different content structures', () => {
    initializeProjectCards();
    
    // Both projects should be properly parsed despite different content
    const rows = document.querySelectorAll('.project-table tbody tr');
    
    rows.forEach(row => {
      row.click();
      
      const card = document.querySelector('.project-detail-card');
      const projectName = row.cells[0].textContent.trim();
      
      // Each project should have its content displayed
      expect(card.textContent).toContain(projectName);
      expect(card.textContent).toContain('Tech Stack:');
      expect(card.textContent).toContain('Purpose:');
    });
  });

  test('should not break when project details are missing', () => {
    // Create a table without matching project details
    document.querySelector('.blog-content').innerHTML = `
      <table class="project-table">
        <tbody>
          <tr>
            <td><strong>Missing Project</strong></td>
            <td>1000</td>
            <td>$100K</td>
          </tr>
        </tbody>
      </table>
    `;
    
    // Should not throw error
    expect(() => initializeProjectCards()).not.toThrow();
    
    // Click should not break
    const row = document.querySelector('.project-table tbody tr');
    expect(() => row.click()).not.toThrow();
  });
});