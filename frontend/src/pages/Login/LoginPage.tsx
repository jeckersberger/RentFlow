import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { LogIn, AlertCircle } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import './LoginPage.scss';

interface LoginFormData {
  email: string;
  password: string;
  tenant_slug: string;
}

export default function LoginPage() {
  const navigate = useNavigate();
  const login = useAuthStore((s) => s.login);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    defaultValues: {
      email: '',
      password: '',
      tenant_slug: 'je-soundulight',
    },
  });

  const onSubmit = async (formData: LoginFormData) => {
    setError(null);
    setIsSubmitting(true);

    try {
      await login(formData.email, formData.password, formData.tenant_slug);
      navigate('/', { replace: true });
    } catch (err: unknown) {
      const message =
        err instanceof Error
          ? err.message
          : typeof err === 'object' && err !== null && 'message' in err
            ? String((err as { message: unknown }).message)
            : 'Login fehlgeschlagen';
      setError(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="login-page">
      <motion.div
        className="login-card"
        initial={{ opacity: 0, y: 24 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, ease: 'easeOut' }}
      >
        <div className="login-card__header">
          <h1 className="login-card__title">CrateDesk</h1>
          <p className="login-card__subtitle">Equipment-Rental Management</p>
        </div>

        {error && (
          <motion.div
            className="login-card__error"
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
          >
            <AlertCircle size={16} />
            <span>{error}</span>
          </motion.div>
        )}

        <form className="login-card__form" onSubmit={handleSubmit(onSubmit)}>
          <div className="form-group">
            <label htmlFor="email">E-Mail</label>
            <input
              id="email"
              type="email"
              autoComplete="email"
              placeholder="name@firma.de"
              {...register('email', { required: 'E-Mail ist erforderlich' })}
            />
            {errors.email && (
              <span className="form-error">{errors.email.message}</span>
            )}
          </div>

          <div className="form-group">
            <label htmlFor="password">Passwort</label>
            <input
              id="password"
              type="password"
              autoComplete="current-password"
              placeholder="Passwort eingeben"
              {...register('password', { required: 'Passwort ist erforderlich' })}
            />
            {errors.password && (
              <span className="form-error">{errors.password.message}</span>
            )}
          </div>

          <div className="form-group">
            <label htmlFor="tenant_slug">Mandant</label>
            <input
              id="tenant_slug"
              type="text"
              placeholder="mandant-slug"
              {...register('tenant_slug', { required: 'Mandant ist erforderlich' })}
            />
            {errors.tenant_slug && (
              <span className="form-error">{errors.tenant_slug.message}</span>
            )}
          </div>

          <button
            type="submit"
            className="login-card__submit"
            disabled={isSubmitting}
          >
            {isSubmitting ? (
              <span className="loading-spinner loading-spinner--small" />
            ) : (
              <>
                <LogIn size={18} />
                <span>Anmelden</span>
              </>
            )}
          </button>
        </form>
      </motion.div>
    </div>
  );
}
