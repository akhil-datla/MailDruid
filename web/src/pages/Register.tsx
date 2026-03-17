import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { register, ApiError } from '../api/client';
import { Mail, Loader2, ArrowRight, Server, Shield } from 'lucide-react';

export default function Register() {
  const [form, setForm] = useState({
    name: '',
    email: '',
    receivingEmail: '',
    password: '',
    domain: '',
    port: 993,
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const update = (field: string) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((prev) => ({ ...prev, [field]: field === 'port' ? Number(e.target.value) : e.target.value }));

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await register(form);
      navigate('/login');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex">
      {/* Left side - branding */}
      <div className="hidden lg:flex lg:w-1/2 gradient-mesh relative overflow-hidden">
        <div className="absolute inset-0 bg-black/10" />
        <div className="relative z-10 flex flex-col justify-between p-12 text-white">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-white/20 backdrop-blur flex items-center justify-center">
              <Mail className="w-5 h-5" />
            </div>
            <span className="text-xl font-bold">MailDruid</span>
          </div>

          <div className="max-w-md space-y-8">
            <div>
              <h1 className="text-4xl font-bold leading-tight mb-4">
                Get started in
                <br />
                under a minute.
              </h1>
              <p className="text-lg text-white/70 leading-relaxed">
                Connect your IMAP email account and start receiving intelligent summaries immediately.
              </p>
            </div>

            <div className="space-y-4">
              <div className="flex items-start gap-3">
                <div className="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center shrink-0 mt-0.5">
                  <Server className="w-4 h-4" />
                </div>
                <div>
                  <p className="font-medium text-sm">Any IMAP Provider</p>
                  <p className="text-sm text-white/60">Gmail, Outlook, Yahoo, or your own mail server.</p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <div className="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center shrink-0 mt-0.5">
                  <Shield className="w-4 h-4" />
                </div>
                <div>
                  <p className="font-medium text-sm">Encrypted Credentials</p>
                  <p className="text-sm text-white/60">Your password is encrypted with AES-256 at rest.</p>
                </div>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-4 text-sm text-white/50">
            <span>Open Source</span>
            <span className="w-1 h-1 rounded-full bg-white/30" />
            <span>Self-Hosted</span>
            <span className="w-1 h-1 rounded-full bg-white/30" />
            <span>Privacy-First</span>
          </div>
        </div>
      </div>

      {/* Right side - register form */}
      <div className="flex-1 flex items-center justify-center p-6 bg-white dark:bg-gray-950">
        <div className="w-full max-w-md animate-in">
          <div className="lg:hidden flex items-center gap-3 mb-10">
            <div className="w-10 h-10 rounded-xl gradient-brand flex items-center justify-center">
              <Mail className="w-5 h-5 text-white" />
            </div>
            <span className="text-xl font-bold text-gray-900 dark:text-white">MailDruid</span>
          </div>

          <div className="mb-8">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Create your account</h2>
            <p className="text-gray-500 dark:text-gray-400 mt-2">Connect your email to start getting summaries.</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            {error && (
              <div className="bg-red-50 dark:bg-red-950/50 text-red-600 dark:text-red-400 text-sm px-4 py-3 rounded-xl border border-red-100 dark:border-red-900/50 animate-in">
                {error}
              </div>
            )}

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Full name</label>
              <input
                type="text"
                value={form.name}
                onChange={update('name')}
                required
                className="w-full px-4 py-3 border border-gray-200 dark:border-gray-800 rounded-xl bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white placeholder-gray-400 focus:ring-2 focus:ring-brand-500/20 focus:border-brand-500 outline-none transition-all duration-200"
                placeholder="Jane Doe"
              />
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">IMAP email</label>
                <input
                  type="email"
                  value={form.email}
                  onChange={update('email')}
                  required
                  className="w-full px-4 py-3 border border-gray-200 dark:border-gray-800 rounded-xl bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white placeholder-gray-400 focus:ring-2 focus:ring-brand-500/20 focus:border-brand-500 outline-none transition-all duration-200"
                  placeholder="you@company.com"
                />
              </div>
              <div className="space-y-1.5">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Summaries sent to</label>
                <input
                  type="email"
                  value={form.receivingEmail}
                  onChange={update('receivingEmail')}
                  required
                  className="w-full px-4 py-3 border border-gray-200 dark:border-gray-800 rounded-xl bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white placeholder-gray-400 focus:ring-2 focus:ring-brand-500/20 focus:border-brand-500 outline-none transition-all duration-200"
                  placeholder="you@gmail.com"
                />
              </div>
            </div>

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">IMAP password</label>
              <input
                type="password"
                value={form.password}
                onChange={update('password')}
                required
                minLength={6}
                className="w-full px-4 py-3 border border-gray-200 dark:border-gray-800 rounded-xl bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white placeholder-gray-400 focus:ring-2 focus:ring-brand-500/20 focus:border-brand-500 outline-none transition-all duration-200"
                placeholder="App password recommended"
              />
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div className="col-span-2 space-y-1.5">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">IMAP server</label>
                <input
                  type="text"
                  value={form.domain}
                  onChange={update('domain')}
                  required
                  className="w-full px-4 py-3 border border-gray-200 dark:border-gray-800 rounded-xl bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white placeholder-gray-400 focus:ring-2 focus:ring-brand-500/20 focus:border-brand-500 outline-none transition-all duration-200"
                  placeholder="imap.gmail.com"
                />
              </div>
              <div className="space-y-1.5">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Port</label>
                <input
                  type="number"
                  value={form.port}
                  onChange={update('port')}
                  required
                  min={1}
                  max={65535}
                  className="w-full px-4 py-3 border border-gray-200 dark:border-gray-800 rounded-xl bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white placeholder-gray-400 focus:ring-2 focus:ring-brand-500/20 focus:border-brand-500 outline-none transition-all duration-200"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 px-4 gradient-brand hover:opacity-90 disabled:opacity-50 text-white font-semibold rounded-xl transition-all duration-200 flex items-center justify-center gap-2 shadow-lg shadow-brand-500/25 cursor-pointer"
            >
              {loading ? (
                <Loader2 className="w-5 h-5 animate-spin" />
              ) : (
                <>
                  Create Account
                  <ArrowRight className="w-4 h-4" />
                </>
              )}
            </button>
          </form>

          <p className="text-center text-sm text-gray-500 dark:text-gray-400 mt-8">
            Already have an account?{' '}
            <Link to="/login" className="text-brand-600 dark:text-brand-400 hover:text-brand-700 dark:hover:text-brand-300 font-semibold transition-colors">
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
