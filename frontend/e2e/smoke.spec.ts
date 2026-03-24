import { test, expect } from '@playwright/test'

test.describe('RentFlow Smoke Tests', () => {
  test('login page loads', async ({ page }) => {
    await page.goto('/login')
    await expect(page.locator('h1', { hasText: 'RentFlow' })).toBeVisible()
    await expect(page.getByLabel('E-Mail-Adresse')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Anmelden' })).toBeVisible()
  })

  test('forgot password page loads', async ({ page }) => {
    await page.goto('/forgot-password')
    await expect(page.locator('h1', { hasText: 'Passwort vergessen' })).toBeVisible()
  })

  test('booking response page handles invalid token', async ({ page }) => {
    await page.goto('/booking/invalid-token')
    await expect(page.getByText('nicht gefunden')).toBeVisible({ timeout: 10_000 })
  })

  test('login and see dashboard', async ({ page }) => {
    await page.goto('/login')
    await page.getByLabel('E-Mail-Adresse').fill('j.eckersberger@je-soundulight.de')
    await page.getByLabel('Passwort').fill('RentFlow2026!')
    await page.getByRole('button', { name: 'Anmelden' }).click()
    await expect(page).toHaveURL('/', { timeout: 10_000 })
    await expect(page.getByText('Dashboard')).toBeVisible()
  })

  test('equipment list loads', async ({ page }) => {
    // Login first
    await page.goto('/login')
    await page.getByLabel('E-Mail-Adresse').fill('j.eckersberger@je-soundulight.de')
    await page.getByLabel('Passwort').fill('RentFlow2026!')
    await page.getByRole('button', { name: 'Anmelden' }).click()
    await expect(page).toHaveURL('/', { timeout: 10_000 })

    await page.goto('/equipment')
    await expect(page.getByText('Ausrüstungsverwaltung')).toBeVisible()
  })

  test('projects list loads', async ({ page }) => {
    // Login first
    await page.goto('/login')
    await page.getByLabel('E-Mail-Adresse').fill('j.eckersberger@je-soundulight.de')
    await page.getByLabel('Passwort').fill('RentFlow2026!')
    await page.getByRole('button', { name: 'Anmelden' }).click()
    await expect(page).toHaveURL('/', { timeout: 10_000 })

    await page.goto('/projects')
    await expect(page.getByText('Projekte')).toBeVisible()
  })
})
