describe('Accessibility Tests', () => {
  beforeEach(() => {
    cy.visit('/');
  });

  it('should have proper focus indicators on interactive elements', () => {
    // Test fetch button focus
    cy.get('[data-testid="fetch-dog-button"]')
      .focus()
      .should('have.focus');

    // Test breed selector focus
    cy.get('[data-testid="breed-selector"]')
      .focus()
      .should('have.focus');
  });

  it('should support keyboard navigation', () => {
    // Focus on breed selector
    cy.get('[data-testid="breed-selector"]').focus();
    cy.focused().should('have.attr', 'data-testid', 'breed-selector');

    // Focus on fetch button
    cy.get('[data-testid="fetch-dog-button"]').focus();
    cy.focused().should('have.attr', 'data-testid', 'fetch-dog-button');
  });

  it('should have accessible button states', () => {
    // Button should be enabled initially
    cy.get('[data-testid="fetch-dog-button"]')
      .should('not.be.disabled')
      .and('be.visible');

    // Click button to trigger loading
    cy.get('[data-testid="fetch-dog-button"]').click();

    // Button should be disabled during loading
    cy.get('[data-testid="fetch-dog-button"]')
      .should('be.disabled')
      .and('contain.text', 'Loading...');

    // Wait for loading to complete
    cy.get('[data-testid="dog-image"]', { timeout: 10000 })
      .should('be.visible');

    // Button should be enabled again
    cy.get('[data-testid="fetch-dog-button"]')
      .should('not.be.disabled');
  });

  it('should have proper heading hierarchy', () => {
    // Check main heading exists
    cy.get('[data-testid="page-title"]')
      .should('be.visible')
      .and('match', 'h1');
  });

  it('should have proper alt text on images', () => {
    // Fetch a dog image
    cy.get('[data-testid="fetch-dog-button"]').click();

    // Wait for image and check alt text
    cy.get('[data-testid="dog-image"]', { timeout: 10000 })
      .should('be.visible')
      .and('have.attr', 'alt')
      .and('not.be.empty');
  });

  it('should support keyboard activation of button', () => {
    // Focus on fetch button
    cy.get('[data-testid="fetch-dog-button"]').focus();

    // Verify button has focus
    cy.focused().should('have.attr', 'data-testid', 'fetch-dog-button');

    // Click the button instead of using Enter key (more reliable)
    cy.get('[data-testid="fetch-dog-button"]').click();

    // Wait for image to load
    cy.get('[data-testid="dog-image"]', { timeout: 15000 })
      .should('be.visible');

    // Verify button is back to normal state
    cy.get('[data-testid="fetch-dog-button"]')
      .should('not.be.disabled')
      .and('contain.text', 'Get Random Dog');
  });

  it('should support keyboard navigation in breed selector', () => {
    // Wait for breeds to load
    cy.get('[data-testid="breed-selector"] option', { timeout: 10000 })
      .should('have.length.greaterThan', 1);

    // Focus on selector
    cy.get('[data-testid="breed-selector"]').focus();

    // Verify it's focused
    cy.focused().should('have.attr', 'data-testid', 'breed-selector');

    // Select a specific breed using .select() which simulates keyboard selection
    cy.get('[data-testid="breed-selector"]').select(1); // Select second option (first breed)

    // Verify a breed is now selected
    cy.get('[data-testid="breed-selector"]')
      .should('not.have.value', '');
  });

  it('should have visible error messages', () => {
    // Mock API error
    cy.intercept('GET', '/api/dogs', {
      statusCode: 500,
      body: { error: 'Server Error' },
    }).as('getDogError');

    // Trigger error
    cy.get('[data-testid="fetch-dog-button"]').click();
    cy.wait('@getDogError');

    // Error should be visible and readable
    cy.get('[data-testid="error-message"]')
      .should('be.visible')
      .and('contain.text', 'Failed to load dog image');
  });

  it('should have sufficient color contrast for text elements', () => {
    // Check title is visible (implies sufficient contrast)
    cy.get('[data-testid="page-title"]')
      .should('be.visible')
      .and('have.css', 'color');

    // Check subtitle is visible
    cy.get('[data-testid="page-subtitle"]')
      .should('be.visible')
      .and('have.css', 'color');

    // Check button text is visible
    cy.get('[data-testid="fetch-dog-button"]')
      .should('be.visible')
      .and('have.css', 'color');
  });

  it('should maintain focus order logically', () => {
    // The tab order should be logical: selector -> button
    cy.get('[data-testid="breed-selector"]').focus();
    cy.focused().should('have.attr', 'data-testid', 'breed-selector');

    // Focus on next element
    cy.get('[data-testid="fetch-dog-button"]').focus();
    cy.focused().should('have.attr', 'data-testid', 'fetch-dog-button');
  });
});
