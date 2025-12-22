import type { Route } from './+types/home';
import { HomeLayout } from 'fumadocs-ui/layouts/home';
import { Link } from 'react-router';
import { baseOptions } from '@/lib/layout.shared';
import { Card, Cards } from '@/components/cards';
import { TOMLHighlight } from '@/components/toml-highlight';

export function meta({}: Route.MetaArgs) {
  return [
    { title: 'Prun' },
    { name: 'description', content: 'Run multiple commands in parallel with real-time output streaming.' },
  ];
}

export default function Home() {
  return (
    <HomeLayout {...baseOptions()}>
      <div className="flex flex-col items-center min-h-[calc(100vh-4rem)]">
        {/* Hero Section */}
        <div className="relative w-full overflow-hidden">
          {/* Background gradient */}
          <div className="absolute inset-0 bg-gradient-to-b from-fd-primary/5 via-transparent to-transparent" />
          
          <div className="relative max-w-4xl mx-auto text-center px-6 py-20 md:py-32">
            <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-fd-primary/10 border border-fd-primary/20 text-sm text-fd-primary mb-6">
              <span className="w-2 h-2 rounded-full bg-fd-primary animate-pulse" />
              CLI Tool for Parallel Task Execution
            </div>
            
            <div className="mb-8 flex justify-center">
              <img 
                src="/Prun.png" 
                alt="Prun Logo" 
                className="h-24 md:h-32 w-auto drop-shadow-lg"
              />
            </div>
            
            <h1 className="text-5xl md:text-7xl font-extrabold mb-6 leading-tight">
              <span className="bg-gradient-to-r from-fd-primary via-fd-primary/80 to-fd-primary/60 bg-clip-text text-transparent">
                Prun
              </span>
            </h1>
            
            <p className="text-2xl md:text-3xl font-semibold text-fd-foreground mb-4">
              Run multiple commands in parallel
            </p>
            <p className="text-lg text-fd-muted-foreground mb-10 max-w-2xl mx-auto">
              A powerful CLI tool that reads a <code className="px-2 py-1 rounded-md bg-fd-muted text-fd-foreground font-mono text-sm">prun.toml</code> configuration file, 
              starts multiple tasks simultaneously, and streams their combined output in real time.
            </p>
            
            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
              <Link
                className="group relative px-8 py-4 bg-fd-primary text-fd-primary-foreground rounded-full font-semibold text-base hover:bg-fd-primary/90 transition-all duration-300 hover:scale-105 hover:shadow-lg hover:shadow-fd-primary/30"
                to="/docs/installation"
              >
                Install
                <span className="inline-block ml-2 transform group-hover:translate-x-1 transition-transform">→</span>
              </Link>
              <Link
                className="px-8 py-4 border-2 border-fd-border rounded-full font-semibold text-base hover:bg-fd-muted hover:border-fd-primary/50 transition-all duration-300"
                to="/docs"
              >
                View Documentation
              </Link>
            </div>

            {/* Code Example Preview */}
            <div className="mt-16 max-w-2xl mx-auto">
              <div className="rounded-xl border border-fd-border bg-fd-card/50 backdrop-blur-sm p-6 text-left">
                <div className="flex items-center gap-2 mb-4">
                  <div className="w-3 h-3 rounded-full bg-red-500" />
                  <div className="w-3 h-3 rounded-full bg-yellow-500" />
                  <div className="w-3 h-3 rounded-full bg-green-500" />
                  <span className="ml-2 text-xs text-fd-muted-foreground font-mono">prun.toml</span>
                </div>
                <TOMLHighlight code={`tasks = ["app", "redis", "server"]

[task.app]
cmd = "npm run dev"
env = { NODE_ENV = "development", PORT = "3000" }

[task.redis]
cmd = "redis-server"

[task.server]
cmd = "go run main.go"
path = "./backend"
env = { PORT = "8080" }
watch = true`} />
              </div>
            </div>
          </div>
        </div>

        {/* Documentation Cards */}
        <div className="w-full max-w-7xl mx-auto px-6 py-16">
          <div className="text-center mb-12">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">Documentation</h2>
            <p className="text-lg text-fd-muted-foreground max-w-2xl mx-auto">
              Everything you need to get started with prun
            </p>
          </div>
          <Cards>
            <Card
              title="Getting Started"
              href="/docs"
              description="Learn how to use prun in your project with step-by-step instructions and examples."
              icon="🚀"
              gradient="bg-gradient-to-br from-blue-500 to-cyan-500"
            />
            <Card
              title="Installation"
              href="/docs/installation"
              description="Install prun on your system. Build from source and set up your development environment."
              icon="📦"
              gradient="bg-gradient-to-br from-purple-500 to-pink-500"
            />
            <Card
              title="Configuration"
              href="/docs/prun-toml-configuration"
              description="Complete reference for configuring prun.toml files with all available options and examples."
              icon="⚙️"
              gradient="bg-gradient-to-br from-orange-500 to-red-500"
            />
          </Cards>
        </div>

        {/* Features Section */}
        <div className="w-full max-w-6xl mx-auto px-6 py-16">
          <div className="text-center mb-12">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">Powerful Features</h2>
            <p className="text-lg text-fd-muted-foreground max-w-2xl mx-auto">
              Everything you need for efficient parallel task management
            </p>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="group p-8 rounded-xl border border-fd-border bg-gradient-to-br from-fd-card to-fd-card/50 hover:border-fd-primary/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
              <div className="text-4xl mb-4">⚡</div>
              <h3 className="font-bold text-xl mb-3 text-fd-foreground">Parallel Execution</h3>
              <p className="text-fd-muted-foreground leading-relaxed">
                Run multiple commands simultaneously with real-time output streaming and task prefixes for easy log identification.
              </p>
            </div>
            <div className="group p-8 rounded-xl border border-fd-border bg-gradient-to-br from-fd-card to-fd-card/50 hover:border-fd-primary/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
              <div className="text-4xl mb-4">🎨</div>
              <h3 className="font-bold text-xl mb-3 text-fd-foreground">Interactive TUI</h3>
              <p className="text-fd-muted-foreground leading-relaxed">
                Beautiful terminal UI with task list, filtered logs, and keyboard navigation for easy monitoring of all your processes.
              </p>
            </div>
            <div className="group p-8 rounded-xl border border-fd-border bg-gradient-to-br from-fd-card to-fd-card/50 hover:border-fd-primary/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
              <div className="text-4xl mb-4">👀</div>
              <h3 className="font-bold text-xl mb-3 text-fd-foreground">File Watching</h3>
              <p className="text-fd-muted-foreground leading-relaxed">
                Automatically restart tasks when files change, perfect for development workflows with hot-reload capabilities.
              </p>
            </div>
            <div className="group p-8 rounded-xl border border-fd-border bg-gradient-to-br from-fd-card to-fd-card/50 hover:border-fd-primary/50 transition-all duration-300 hover:shadow-lg hover:-translate-y-1">
              <div className="text-4xl mb-4">🛡️</div>
              <h3 className="font-bold text-xl mb-3 text-fd-foreground">Graceful Shutdown</h3>
              <p className="text-fd-muted-foreground leading-relaxed">
                Clean signal handling with automatic cleanup when tasks fail or are interrupted, ensuring no orphaned processes.
              </p>
            </div>
          </div>
        </div>
      </div>
    </HomeLayout>
  );
}
