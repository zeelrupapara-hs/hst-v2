import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import HttpApi from 'i18next-http-backend';
import LanguageDetector from 'i18next-browser-languagedetector';

i18n
  .use(HttpApi) // Load translation files via backend
  .use(LanguageDetector) // Detect user language
  .use(initReactI18next) // Initialize with React
  .init({
    fallbackLng: 'en',
    supportedLngs: ['en', 'ar', 'hi'],
    debug: true,
    backend: {
      loadPath: '/locales/{{lng}}.json', // Translation file path
    },
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      caches: ['localStorage'], // Save language preference to localStorage
    },
    interpolation: {
      escapeValue: false, // React already escapes by default
    },
    react: {
      useSuspense: true, // Enables Suspense
    },
  });

export default i18n;
