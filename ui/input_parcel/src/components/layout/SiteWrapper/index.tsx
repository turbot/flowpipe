import { ReactNode } from "react";

interface SiteWrapperProps {
  children: ReactNode
}

const SiteWrapper = ({children}:SiteWrapperProps) => {
  return (<main className="flex min-h-screen flex-col items-center justify-between p-24">
    {children}
  </main>)
};

export default SiteWrapper