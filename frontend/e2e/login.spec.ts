import { test, expect } from '@playwright/test';

test('login and navigate to equipment', async ({ page }) => {
  await page.goto('http://46.224.105.10/login');
  await page.fill('[placeholder="name@firma.de"]', 'j.eckersberger@je-soundulight.de');
  await page.fill('[placeholder="Passwort eingeben"]', 'RentFlow2026!');
  await page.click('button:has-text("Anmelden")');
  await expect(page).toHaveURL('http://46.224.105.10/');
  await expect(page.locator('h2')).toContainText('Dashboard');

  // Navigate to equipment
  await page.click('a[href="/equipment"]');
  await expect(page.locator('h2')).toContainText('Equipment');
  await expect(page.locator('table')).toBeVisible();
});

test('create new customer', async ({ page }) => {
  // Login first
  await page.goto('http://46.224.105.10/login');
  await page.fill('[placeholder="name@firma.de"]', 'j.eckersberger@je-soundulight.de');
  await page.fill('[placeholder="Passwort eingeben"]', 'RentFlow2026!');
  await page.click('button:has-text("Anmelden")');

  // Go to customers
  await page.click('a[href="/customers"]');
  await page.click('button:has-text("Neu")');
  await expect(page).toHaveURL(/customers\/new/);
});
