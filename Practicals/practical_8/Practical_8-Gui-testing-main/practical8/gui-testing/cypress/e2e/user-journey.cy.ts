describe('Complete User Journey', () => {
  it('should complete a full user workflow from start to finish', () => {
    // 1. User visits homepage
    cy.visit('/');

    // 2. User sees welcome message
    cy.get('[data-testid="page-title"]')
      .should('be.visible')
      .and('contain.text', 'Dog Image Browser');

    cy.get('[data-testid="page-subtitle"]')
      .should('be.visible')
      .and('contain.text', 'Powered by Dog CEO API');

    // 3. User sees placeholder message
    cy.get('[data-testid="placeholder-message"]')
      .should('be.visible')
      .and('contain.text', 'Click "Get Random Dog" to see a cute dog!');

    // 4. User browses available breeds - wait for breeds to load
    cy.get('[data-testid="breed-selector"] option', { timeout: 10000 })
      .should('have.length.greaterThan', 1);

    // Verify first option is "All Breeds"
    cy.get('[data-testid="breed-selector"] option')
      .first()
      .should('have.text', 'All Breeds (Random)');

    // 5. User selects a specific breed (husky)
    cy.get('[data-testid="breed-selector"]').select('husky');
    cy.get('[data-testid="breed-selector"]').should('have.value', 'husky');

    // 6. User fetches dog image
    cy.get('[data-testid="fetch-dog-button"]').click();

    // 7. User sees loading state
    cy.get('[data-testid="fetch-dog-button"]')
      .should('contain.text', 'Loading...')
      .and('be.disabled');

    // 8. User views the image - wait for it to load
    cy.get('[data-testid="dog-image"]', { timeout: 10000 })
      .should('be.visible')
      .and('have.attr', 'src')
      .and('include', 'husky');

    // Placeholder should be gone
    cy.get('[data-testid="placeholder-message"]').should('not.exist');

    // 9. User selects different breed (corgi)
    cy.get('[data-testid="breed-selector"]').select('corgi');
    cy.get('[data-testid="breed-selector"]').should('have.value', 'corgi');

    // 10. User fetches another image
    cy.get('[data-testid="fetch-dog-button"]').click();
    cy.get('[data-testid="dog-image"]', { timeout: 10000 })
      .should('be.visible')
      .and('have.attr', 'src')
      .and('include', 'corgi');

    // 11. User selects "All Breeds"
    cy.get('[data-testid="breed-selector"]').select('');
    cy.get('[data-testid="breed-selector"]').should('have.value', '');

    // 12. User fetches random dog
    cy.get('[data-testid="fetch-dog-button"]').click();
    cy.get('[data-testid="dog-image"]', { timeout: 10000 })
      .should('be.visible')
      .and('have.attr', 'src')
      .and('include', 'images.dog.ceo');

    // Verify no errors occurred
    cy.get('[data-testid="error-message"]').should('not.exist');
  });

  it('should handle error and recovery in user journey', () => {
    // 1. User visits homepage
    cy.visit('/');

    // 2. Mock API failure
    cy.intercept('GET', '/api/dogs', {
      statusCode: 500,
      body: { error: 'Server Error' },
    }).as('getDogError');

    // 3. User tries to fetch dog
    cy.get('[data-testid="fetch-dog-button"]').click();
    cy.wait('@getDogError');

    // 4. User sees error message
    cy.get('[data-testid="error-message"]')
      .should('be.visible')
      .and('contain.text', 'Failed to load dog image');

    // 5. User tries again with successful API
    cy.intercept('GET', '/api/dogs', {
      statusCode: 200,
      body: {
        message: 'https://images.dog.ceo/breeds/husky/n02110185_1469.jpg',
        status: 'success',
      },
    }).as('getDogSuccess');

    // 6. User clicks fetch button again
    cy.get('[data-testid="fetch-dog-button"]').click();
    cy.wait('@getDogSuccess');

    // 7. Error message should disappear
    cy.get('[data-testid="error-message"]').should('not.exist');

    // 8. Dog image should appear
    cy.get('[data-testid="dog-image"]')
      .should('be.visible')
      .and('have.attr', 'src')
      .and('include', 'n02110185_1469.jpg');
  });

  it('should allow multiple breed selections in succession', () => {
    cy.visit('/');

    // Wait for breeds to load
    cy.get('[data-testid="breed-selector"] option', { timeout: 10000 })
      .should('have.length.greaterThan', 1);

    const breeds = ['husky', 'poodle', 'retriever'];

    for (const breed of breeds) {
      // Select breed
      cy.get('[data-testid="breed-selector"]').select(breed);

      // Fetch dog
      cy.get('[data-testid="fetch-dog-button"]').click();

      // Verify correct breed image loaded
      cy.get('[data-testid="dog-image"]', { timeout: 10000 })
        .should('be.visible')
        .and('have.attr', 'src')
        .and('include', breed);
    }
  });
});
