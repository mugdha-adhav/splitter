export default function Home() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center gap-8 p-8">
      <h1 className="text-4xl font-bold">Splitter</h1>

      <div className="flex gap-4">
        <a href="/auth/login" className="px-6 py-2 rounded-full bg-foreground text-background hover:bg-[#383838] dark:hover:bg-[#ccc] transition-colors">
          Sign In
        </a>
        <a href="/auth/register" className="px-6 py-2 rounded-full border border-black/[.08] dark:border-white/[.145] hover:bg-[#f2f2f2] dark:hover:bg-[#1a1a1a] transition-colors">
          Sign Up
        </a>
      </div>
    </div>
  );
}
