import type { JSX } from "react";

interface TOMLHighlightProps {
  code: string;
}

export function TOMLHighlight({ code }: TOMLHighlightProps) {
  const highlightTOML = (text: string) => {
    const lines = text.split('\n');
    
    return lines.map((line, lineIndex) => {
      // Handle empty lines
      if (!line.trim()) {
        return <span key={lineIndex} className="block">{'\n'}</span>;
      }

      const parts: JSX.Element[] = [];
      let remaining = line;

      // Match section headers [task.app]
      const sectionMatch = remaining.match(/^(\s*)(\[[^\]]+\])/);
      if (sectionMatch) {
        parts.push(
          <span key="indent">{sectionMatch[1]}</span>,
          <span key="section" className="text-purple-400 dark:text-purple-300">{sectionMatch[2]}</span>
        );
        return (
          <span key={lineIndex} className="block">
            {parts}
          </span>
        );
      }

      // Match key = value patterns
      const keyValueMatch = remaining.match(/^(\s*)([a-zA-Z_][a-zA-Z0-9_-]*)\s*=\s*(.*)$/);
      if (keyValueMatch) {
        const indent = keyValueMatch[1];
        const key = keyValueMatch[2];
        const value = keyValueMatch[3].trim();
        
        parts.push(
          <span key="indent">{indent}</span>,
          <span key="key" className="text-blue-400 dark:text-blue-300">{key}</span>,
          <span key="equals" className="text-fd-muted-foreground"> = </span>
        );
        
        // Highlight the value
        const highlightedValue = highlightValue(value);
        parts.push(...highlightedValue);
      } else {
        // Plain text line
        parts.push(<span key="plain">{remaining}</span>);
      }

      return (
        <span key={lineIndex} className="block">
          {parts}
        </span>
      );
    });
  };

  const highlightValue = (value: string): JSX.Element[] => {
    const parts: JSX.Element[] = [];
    let remaining = value.trim();
    let key = 0;

    // Match arrays ["app", "redis", "server"]
    const arrayMatch = remaining.match(/^\[(.*?)\]$/);
    if (arrayMatch) {
      parts.push(
        <span key={key++} className="text-yellow-400 dark:text-yellow-300">[</span>
      );
      
      // Split array items by comma, but preserve quotes
      const arrayContent = arrayMatch[1];
      if (arrayContent.trim()) {
        // Simple split that handles quoted strings
        const items: string[] = [];
        let current = '';
        let inQuotes = false;
        
        for (let i = 0; i < arrayContent.length; i++) {
          const char = arrayContent[i];
          if (char === '"' && (i === 0 || arrayContent[i - 1] !== '\\')) {
            inQuotes = !inQuotes;
            current += char;
          } else if (char === ',' && !inQuotes) {
            if (current.trim()) items.push(current.trim());
            current = '';
          } else {
            current += char;
          }
        }
        if (current.trim()) items.push(current.trim());
        
        items.forEach((item, index) => {
          if (item.startsWith('"') && item.endsWith('"')) {
            parts.push(
              <span key={key++} className="text-green-400 dark:text-green-300">{item}</span>
            );
          } else {
            parts.push(<span key={key++}>{item}</span>);
          }
          if (index < items.length - 1) {
            parts.push(<span key={key++} className="text-fd-muted-foreground">, </span>);
          }
        });
      }
      
      parts.push(
        <span key={key++} className="text-yellow-400 dark:text-yellow-300">]</span>
      );
      return parts;
    }

    // Match inline tables { KEY = "value", KEY2 = "value2" }
    const inlineTableMatch = remaining.match(/^\{([^}]*)\}$/);
    if (inlineTableMatch) {
      parts.push(
        <span key={key++} className="text-yellow-400 dark:text-yellow-300">{'{'}</span>
      );
      
      const tableContent = inlineTableMatch[1].trim();
      if (tableContent) {
        // Parse key-value pairs manually to handle quoted strings correctly
        const pairs: Array<{ key: string; value: string }> = [];
        let inQuotes = false;
        let currentPair = '';
        
        // First, split by comma while respecting quoted strings
        for (let i = 0; i < tableContent.length; i++) {
          const char = tableContent[i];
          if (char === '"' && (i === 0 || tableContent[i - 1] !== '\\')) {
            inQuotes = !inQuotes;
            currentPair += char;
          } else if (char === ',' && !inQuotes) {
            if (currentPair.trim()) {
              // Parse this pair: KEY = VALUE
              const pairStr = currentPair.trim();
              const eqIndex = pairStr.indexOf('=');
              if (eqIndex > 0) {
                const k = pairStr.substring(0, eqIndex).trim();
                const v = pairStr.substring(eqIndex + 1).trim();
                pairs.push({ key: k, value: v });
              }
            }
            currentPair = '';
          } else {
            currentPair += char;
          }
        }
        
        // Handle the last pair
        if (currentPair.trim()) {
          const pairStr = currentPair.trim();
          const eqIndex = pairStr.indexOf('=');
          if (eqIndex > 0) {
            const k = pairStr.substring(0, eqIndex).trim();
            const v = pairStr.substring(eqIndex + 1).trim();
            pairs.push({ key: k, value: v });
          }
        }
        
        pairs.forEach((pair, index) => {
          parts.push(
            <span key={key++} className="text-blue-400 dark:text-blue-300">{pair.key}</span>,
            <span key={key++} className="text-fd-muted-foreground"> = </span>
          );
          
          const val = pair.value;
          if (val.startsWith('"') && val.endsWith('"')) {
            parts.push(
              <span key={key++} className="text-green-400 dark:text-green-300">{val}</span>
            );
          } else if (val === 'true' || val === 'false') {
            parts.push(
              <span key={key++} className="text-orange-400 dark:text-orange-300">{val}</span>
            );
          } else if (/^\d+$/.test(val)) {
            parts.push(
              <span key={key++} className="text-cyan-400 dark:text-cyan-300">{val}</span>
            );
          } else {
            parts.push(<span key={key++}>{val}</span>);
          }
          
          if (index < pairs.length - 1) {
            parts.push(<span key={key++} className="text-fd-muted-foreground">, </span>);
          }
        });
      }
      
      parts.push(
        <span key={key++} className="text-yellow-400 dark:text-yellow-300">{'}'}</span>
      );
      return parts;
    }

    // Match strings (including paths)
    const stringMatch = remaining.match(/^(".*?")/);
    if (stringMatch) {
      parts.push(
        <span key={key++} className="text-green-400 dark:text-green-300">{stringMatch[1]}</span>
      );
      return parts;
    }

    // Match booleans
    const boolMatch = remaining.match(/^(true|false)$/);
    if (boolMatch) {
      parts.push(
        <span key={key++} className="text-orange-400 dark:text-orange-300">{boolMatch[1]}</span>
      );
      return parts;
    }

    // Match numbers
    const numberMatch = remaining.match(/^(\d+)$/);
    if (numberMatch) {
      parts.push(
        <span key={key++} className="text-cyan-400 dark:text-cyan-300">{numberMatch[1]}</span>
      );
      return parts;
    }

    // Plain string value (like paths without quotes)
    parts.push(<span key={key++}>{remaining}</span>);
    return parts;
  };

  return (
    <pre className="text-sm text-fd-foreground font-mono overflow-x-auto">
      <code>{highlightTOML(code)}</code>
    </pre>
  );
}

