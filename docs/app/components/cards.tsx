import { Link } from 'react-router';

interface CardProps {
  title: string;
  href: string;
  description: string;
  icon?: React.ReactNode;
  gradient?: string;
}

export function Card({ title, href, description, icon, gradient }: CardProps) {
  return (
    <Link
      to={href}
      className="group relative flex flex-col gap-3 rounded-xl border border-fd-border bg-fd-card p-6 transition-all duration-300 hover:border-fd-primary/50 hover:shadow-lg hover:shadow-fd-primary/10 hover:-translate-y-1 overflow-hidden"
    >
      {/* Gradient background on hover */}
      {gradient && (
        <div className={`absolute inset-0 opacity-0 group-hover:opacity-5 transition-opacity duration-300 ${gradient}`} />
      )}
      
      <div className="relative z-10">
        {icon && (
          <div className="mb-3 text-3xl transform group-hover:scale-110 transition-transform duration-300">
            {icon}
          </div>
        )}
        <h3 className="font-bold text-lg text-fd-foreground group-hover:text-fd-primary transition-colors mb-2">
          {title}
        </h3>
        <p className="text-sm text-fd-muted-foreground leading-relaxed">{description}</p>
        <div className="mt-4 pt-2 text-sm font-medium text-fd-primary opacity-0 group-hover:opacity-100 transition-all duration-300 flex items-center gap-1">
          Learn more
          <span className="transform group-hover:translate-x-1 transition-transform">→</span>
        </div>
      </div>
    </Link>
  );
}

interface CardsProps {
  children: React.ReactNode;
}

export function Cards({ children }: CardsProps) {
  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3 mt-8">
      {children}
    </div>
  );
}

