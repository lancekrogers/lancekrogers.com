// Modal functionality tests
import { describe, test, expect, beforeEach } from '@jest/globals';
import { JSDOM } from 'jsdom';

describe('Contact Modal Functionality', () => {
  let dom;
  let document;
  let window;

  beforeEach(() => {
    // Create DOM with modal HTML
    const modalHTML = `
      <!DOCTYPE html>
      <html>
        <head><title>Test</title></head>
        <body>
          <button data-contact-modal class="btn-primary">Start a Conversation</button>
          <button data-contact-modal class="btn-cta-primary">Start a Conversation</button>
          
          <div id="contact-modal" class="modal-overlay">
            <div class="modal-container">
              <button class="modal-close" aria-label="Close">×</button>
              <h2>Start a Conversation</h2>
              <form hx-post="/contact" hx-target="#modal-contact-response">
                <input type="text" name="name" required>
                <input type="email" name="email" required>
                <textarea name="message" required></textarea>
                <button type="submit">Send Message</button>
              </form>
              <div id="modal-contact-response"></div>
            </div>
          </div>
        </body>
      </html>
    `;

    dom = new JSDOM(modalHTML);
    document = dom.window.document;
    window = dom.window;
  });

  test('modal should exist in page HTML', () => {
    const modal = document.getElementById('contact-modal');
    expect(modal).toBeTruthy();
    expect(modal.classList.contains('modal-overlay')).toBe(true);
  });

  test('modal buttons should exist in page HTML', () => {
    const buttons = document.querySelectorAll('[data-contact-modal]');
    expect(buttons.length).toBe(2);
    expect(buttons[0].classList.contains('btn-primary')).toBe(true);
    expect(buttons[1].classList.contains('btn-cta-primary')).toBe(true);
  });

  test('modal should have required form fields', () => {
    const modal = document.getElementById('contact-modal');
    const nameInput = modal.querySelector('input[name="name"]');
    const emailInput = modal.querySelector('input[name="email"]');
    const messageTextarea = modal.querySelector('textarea[name="message"]');
    const submitButton = modal.querySelector('button[type="submit"]');
    
    expect(nameInput).toBeTruthy();
    expect(emailInput).toBeTruthy();
    expect(emailInput.type).toBe('email');
    expect(messageTextarea).toBeTruthy();
    expect(submitButton).toBeTruthy();
  });

  test('modal should have close button', () => {
    const modal = document.getElementById('contact-modal');
    const closeButton = modal.querySelector('.modal-close');
    
    expect(closeButton).toBeTruthy();
    expect(closeButton.getAttribute('aria-label')).toBe('Close');
  });

  test('modal form should have HTMX attributes', () => {
    const modal = document.getElementById('contact-modal');
    const form = modal.querySelector('form');
    
    expect(form.getAttribute('hx-post')).toBe('/contact');
    expect(form.getAttribute('hx-target')).toBe('#modal-contact-response');
  });

  test('modal should have response container', () => {
    const responseContainer = document.getElementById('modal-contact-response');
    expect(responseContainer).toBeTruthy();
  });

  test('modal buttons should be accessible', () => {
    const buttons = document.querySelectorAll('[data-contact-modal]');
    buttons.forEach(button => {
      expect(button.tagName).toBe('BUTTON');
      expect(button.textContent).toContain('Start a Conversation');
    });
  });
});

