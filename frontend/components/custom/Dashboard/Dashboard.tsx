import { Navbar } from "@/components/custom/Navbar/Navbar";

export default function Dashboard({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen">
      <Navbar />
      {children}
    </div>
  );
}
