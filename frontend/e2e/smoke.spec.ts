import { test, expect, type Page } from '@playwright/test'

const LOGIN_USER = process.env.E2E_LOGIN_USER
const LOGIN_PASS = process.env.E2E_LOGIN_PASS

async function login(page: Page) {
  if (!LOGIN_USER || !LOGIN_PASS) {
    throw new Error('E2E_LOGIN_USER and E2E_LOGIN_PASS must be set for authenticated smoke tests')
  }

  await page.goto('/login')
  await page.getByLabel('Benutzername oder E-Mail').fill(LOGIN_USER)
  await page.getByLabel('Passwort').fill(LOGIN_PASS)
  await page.getByRole('button', { name: 'Anmelden' }).click()
  await expect(page).toHaveURL('/', { timeout: 15_000 })
}

test.describe('CrateDesk smoke tests', () => {
  test('login page loads', async ({ page }) => {
    await page.goto('/login')
    await expect(page.getByLabel('Benutzername oder E-Mail')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Anmelden' })).toBeVisible()
  })

  test('login succeeds', async ({ page }) => {
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

  test('forgot password page loads', async ({ page }) => {
    await page.goto('/forgot-password')
    await expect(page.locator('h1')).toBeVisible()
  })

  test('impressum page loads', async ({ page }) => {
    await page.goto('/impressum')
    await expect(page.locator('body')).toContainText('Impressum')
  })
})
