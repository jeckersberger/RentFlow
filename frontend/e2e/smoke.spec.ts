import { test, expect } from '@playwright/test'

const LOGIN_USER = 'jeck'
const LOGIN_PASS = 'RentFlow2026!'

async function login(page: any) {
  await page.goto('/login')
  await page.getByLabel('Benutzername oder E-Mail').fill(LOGIN_USER)
  await page.getByLabel('Passwort').fill(LOGIN_PASS)
  await page.getByRole('button', { name: 'Anmelden' }).click()
  await expect(page).toHaveURL('/', { timeout: 15_000 })
}

test.describe('RentFlow Smoke Tests', () => {
  test('login page loads', async ({ page }) => {
    await page.goto('/login')
    await expect(page.locator('h1', { hasText: 'RentFlow' })).toBeVisible()
    await expect(page.getByLabel('Benutzername oder E-Mail')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Anmelden' })).toBeVisible()
  })

  test('login with username', async ({ page }) => {
    await login(page)
    await expect(page.getByText('Dashboard')).toBeVisible()
  })

  test('equipment page loads', async ({ page }) => {
    await login(page)
    await page.goto('/equipment')
    await expect(page.getByText('Equipment')).toBeVisible()
  })

  test('projects page loads', async ({ page }) => {
    await login(page)
    await page.goto('/projects')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('h1')).toBeVisible()
  })

  test('invoices page loads', async ({ page }) => {
    await login(page)
    await page.goto('/invoices')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('h1')).toBeVisible()
  })

  test('scanner page loads', async ({ page }) => {
    await login(page)
    await page.goto('/scanner')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('select, input')).toBeVisible()
  })

  test('settings page loads', async ({ page }) => {
    await login(page)
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('h1')).toBeVisible()
  })

  test('create equipment type', async ({ page }) => {
    await login(page)
    await page.goto('/equipment')
    await page.getByRole('button', { name: /Typ erstellen/i }).click()
    await expect(page.getByText('Neues Equipment')).toBeVisible()
    await page.getByPlaceholder(/QSC K12|Name/i).fill('Test-Lautsprecher E2E')
    // Don't actually save — just verify the form opens
  })

  test('create project', async ({ page }) => {
    await login(page)
    await page.goto('/projects')
    await page.getByRole('button', { name: /Neues Projekt|Projekt erstellen/i }).click()
    await page.waitForLoadState('networkidle')
    // Verify form is visible
    await expect(page.locator('input')).toBeVisible()
  })

  test('forgot password page loads', async ({ page }) => {
    await page.goto('/forgot-password')
    await expect(page.locator('h1')).toBeVisible()
  })

  test('impressum page loads', async ({ page }) => {
    await page.goto('/impressum')
    await expect(page.locator('body')).toContainText('Impressum')
  })
})
