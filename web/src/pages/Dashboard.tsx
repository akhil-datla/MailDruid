import { useState, useEffect, useCallback } from 'react';
import { getProfile, generateSummary, type UserProfile, type SummaryResult, ApiError } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { useNavigate, Link } from 'react-router-dom';
import {
  Sparkles,
  Loader2,
  AlertCircle,
  Tag,
  Clock,
  FolderOpen,
  Wand2,
  ArrowRight,
  Mail,
  ImageIcon,
} from 'lucide-react';

export default function Dashboard() {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [summary, setSummary] = useState<SummaryResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState('');
  const { logout } = useAuth();
  const navigate = useNavigate();

  const loadProfile = useCallback(async () => {
    try {
      const data = await getProfile();
      setProfile(data);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        logout();
        navigate('/login');
      }
    } finally {
      setLoading(false);
    }
  }, [logout, navigate]);

  useEffect(() => {
    loadProfile();
  }, [loadProfile]);

  const handleGenerate = async () => {
    setError('');
    setSummary(null);
    setGenerating(true);

    try {
      const result = await generateSummary();
      setSummary(result);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Generation failed');
    } finally {
      setGenerating(false);
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center py-32 gap-3">
        <Loader2 className="w-8 h-8 animate-spin text-brand-500" />
        <span className="text-sm text-gray-400">Loading your dashboard...</span>
      </div>
    );
  }

  const hasTags = profile?.tags && profile.tags.length > 0;
  const hasFolder = !!profile?.folder;
  const hasSchedule = profile?.updateInterval && profile.updateInterval !== '0';

  return (
    <div className="space-y-8 animate-in">
      {/* Welcome */}
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white tracking-tight">
            Welcome back, <span className="text-gradient">{profile?.name}</span>
          </h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1.5">
            Here's an overview of your email summarization setup.
          </p>
        </div>
      </div>

      {/* Status cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 animate-in-delay">
        <StatusCard
          icon={<Tag className="w-5 h-5" />}
          label="Tags"
          value={hasTags ? profile!.tags!.join(', ') : 'Not configured'}
          configured={!!hasTags}
          color="indigo"
        />
        <StatusCard
          icon={<FolderOpen className="w-5 h-5" />}
          label="Folder"
          value={hasFolder ? profile!.folder : 'Not set'}
          configured={hasFolder}
          color="violet"
        />
        <StatusCard
          icon={<Clock className="w-5 h-5" />}
          label="Schedule"
          value={hasSchedule ? `Every ${profile!.updateInterval} min` : 'Not scheduled'}
          configured={!!hasSchedule}
          color="purple"
        />
      </div>

      {/* Setup prompt if not configured */}
      {!hasTags && (
        <div className="animate-in-delay-2 bg-gradient-to-br from-brand-50 to-violet-50 dark:from-brand-950/30 dark:to-violet-950/30 rounded-2xl border border-brand-100 dark:border-brand-900/30 p-8 text-center">
          <div className="w-14 h-14 rounded-2xl gradient-brand flex items-center justify-center mx-auto mb-4 shadow-lg shadow-brand-500/20">
            <Mail className="w-7 h-7 text-white" />
          </div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">Set up your first summary</h3>
          <p className="text-gray-500 dark:text-gray-400 mb-6 max-w-md mx-auto">
            Configure email tags in Settings to tell MailDruid which emails to summarize. You'll be generating summaries in no time.
          </p>
          <Link
            to="/settings"
            className="inline-flex items-center gap-2 px-6 py-2.5 gradient-brand text-white font-semibold rounded-xl shadow-lg shadow-brand-500/25 hover:opacity-90 transition-all"
          >
            Go to Settings
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
      )}

      {/* Generate section */}
      <div className="animate-in-delay-2 bg-white dark:bg-gray-900 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div className="p-6 sm:p-8 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-gray-100 dark:border-gray-800">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-2xl bg-gradient-to-br from-brand-100 to-violet-100 dark:from-brand-900/40 dark:to-violet-900/40 flex items-center justify-center">
              <Wand2 className="w-6 h-6 text-brand-600 dark:text-brand-400" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Generate Summary</h2>
              <p className="text-sm text-gray-500 dark:text-gray-400">
                Summarize your latest emails matching your tags.
              </p>
            </div>
          </div>
          <button
            onClick={handleGenerate}
            disabled={generating || !hasTags}
            className="flex items-center gap-2.5 px-6 py-3 gradient-brand hover:opacity-90 disabled:opacity-40 disabled:cursor-not-allowed text-white font-semibold rounded-xl transition-all duration-200 shadow-lg shadow-brand-500/25 cursor-pointer shrink-0"
          >
            {generating ? (
              <>
                <Loader2 className="w-5 h-5 animate-spin" />
                Generating...
              </>
            ) : (
              <>
                <Sparkles className="w-5 h-5" />
                Generate
              </>
            )}
          </button>
        </div>

        <div className="p-6 sm:p-8">
          {error && (
            <div className="flex items-start gap-3 bg-red-50 dark:bg-red-950/30 text-red-600 dark:text-red-400 text-sm px-5 py-4 rounded-xl border border-red-100 dark:border-red-900/30 mb-6 animate-in">
              <AlertCircle className="w-5 h-5 shrink-0 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {generating && (
            <div className="flex flex-col items-center justify-center py-16 gap-4">
              <div className="relative">
                <div className="w-16 h-16 rounded-full border-4 border-brand-100 dark:border-brand-900/50" />
                <div className="absolute inset-0 w-16 h-16 rounded-full border-4 border-transparent border-t-brand-500 animate-spin" />
              </div>
              <div className="text-center">
                <p className="font-medium text-gray-900 dark:text-white">Processing your emails...</p>
                <p className="text-sm text-gray-400 mt-1">This may take a moment.</p>
              </div>
            </div>
          )}

          {summary && !generating && (
            <div className="space-y-6 animate-in">
              <div className="bg-gray-50 dark:bg-gray-800/50 rounded-xl p-6 border border-gray-100 dark:border-gray-800">
                <div className="flex items-center gap-2 mb-3">
                  <Sparkles className="w-4 h-4 text-brand-500" />
                  <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wider">Summary</h3>
                </div>
                <p className="text-gray-800 dark:text-gray-200 whitespace-pre-wrap leading-relaxed text-[15px]">
                  {summary.summary}
                </p>
              </div>

              {summary.image && (
                <div className="bg-gray-50 dark:bg-gray-800/50 rounded-xl p-6 border border-gray-100 dark:border-gray-800">
                  <div className="flex items-center gap-2 mb-3">
                    <ImageIcon className="w-4 h-4 text-violet-500" />
                    <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wider">Word Cloud</h3>
                  </div>
                  <img
                    src={`data:image/png;base64,${summary.image}`}
                    alt="Word cloud visualization"
                    className="w-full max-w-xl mx-auto rounded-xl shadow-sm"
                  />
                </div>
              )}
            </div>
          )}

          {!summary && !error && !generating && (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <div className="w-16 h-16 rounded-2xl bg-gray-100 dark:bg-gray-800 flex items-center justify-center mb-4">
                <Wand2 className="w-7 h-7 text-gray-300 dark:text-gray-600" />
              </div>
              <p className="text-gray-400 dark:text-gray-500 font-medium">
                {hasTags ? 'Click Generate to summarize your emails.' : 'Configure tags in Settings to get started.'}
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function StatusCard({
  icon,
  label,
  value,
  configured,
  color,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  configured: boolean;
  color: 'indigo' | 'violet' | 'purple';
}) {
  const colorMap = {
    indigo: {
      bg: 'bg-indigo-50 dark:bg-indigo-950/30',
      icon: 'text-indigo-500 dark:text-indigo-400',
      border: 'border-indigo-100 dark:border-indigo-900/30',
    },
    violet: {
      bg: 'bg-violet-50 dark:bg-violet-950/30',
      icon: 'text-violet-500 dark:text-violet-400',
      border: 'border-violet-100 dark:border-violet-900/30',
    },
    purple: {
      bg: 'bg-purple-50 dark:bg-purple-950/30',
      icon: 'text-purple-500 dark:text-purple-400',
      border: 'border-purple-100 dark:border-purple-900/30',
    },
  };

  const colors = configured ? colorMap[color] : {
    bg: 'bg-gray-50 dark:bg-gray-800/50',
    icon: 'text-gray-300 dark:text-gray-600',
    border: 'border-gray-100 dark:border-gray-800',
  };

  return (
    <div className={`rounded-2xl border ${colors.border} ${colors.bg} p-5 transition-all duration-200 hover:shadow-md`}>
      <div className="flex items-center gap-2.5 mb-2">
        <span className={colors.icon}>{icon}</span>
        <span className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">{label}</span>
      </div>
      <p className={`text-sm truncate font-medium ${configured ? 'text-gray-900 dark:text-white' : 'text-gray-400 dark:text-gray-500 italic'}`}>
        {value}
      </p>
    </div>
  );
}
