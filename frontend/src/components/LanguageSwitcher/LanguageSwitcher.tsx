import { useTranslation } from 'react-i18next'
import './LanguageSwitcher.scss'

export function LanguageSwitcher() {
  const { i18n } = useTranslation()

  const handleLanguageChange = (lang: string) => {
    i18n.changeLanguage(lang)
    localStorage.setItem('language', lang)
  }

  return (
    <div className="language-switcher">
      <button
        className={`language-btn ${i18n.language === 'en' ? 'active' : ''}`}
        onClick={() => handleLanguageChange('en')}
        title="English"
      >
        EN
      </button>
      <button
        className={`language-btn ${i18n.language === 'de' ? 'active' : ''}`}
        onClick={() => handleLanguageChange('de')}
        title="Deutsch"
      >
        DE
      </button>
    </div>
  )
}

export default LanguageSwitcher
