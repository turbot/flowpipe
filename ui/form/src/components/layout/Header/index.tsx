import ThemeToggle from "@flowpipe/components/layout/ThemeToggle";

const Header = () => {
  return (
    <div className="flex w-screen items-center justify-end p-4">
      {/*<FlowpipeLogo />*/}
      <ThemeToggle />
    </div>
  );
};

export default Header;
