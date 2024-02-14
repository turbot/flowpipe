import FlowpipeLogo from "components/layout/FlowpipeLogo";
import ThemeToggle from "components/layout/ThemeToggle";

const Header = () => {
  return (
    <div className="flex w-screen items-center justify-between p-4 bg-white">
      <FlowpipeLogo />
      <ThemeToggle />
    </div>
  );
};

export default Header;
