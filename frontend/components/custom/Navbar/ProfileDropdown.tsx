"use client";

import { useUser } from "@/components/custom/Provider/UserProvider";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Key, LogOut, User, UserCircle } from "lucide-react";
import { useState } from "react";

import { PasswordModal } from "../Modal/PasswordModal";
import { ProfileModal } from "../Modal/ProfileModal";

export function ProfileDropdown() {
  const { read: user, logout } = useUser();
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const [isPasswordOpen, setIsPasswordOpen] = useState(false);

  const getInitials = (name: string | null | undefined) => {
    if (!name) return "PN";
    return name
      .split(" ")
      .map((word) => word[0])
      .join("")
      .toUpperCase();
  };

  return (
    <>
      <Popover>
        <PopoverTrigger asChild>
          <Button variant="ghost" className="relative h-8 w-8 rounded-full">
            <Avatar className="h-8 w-8">
              <AvatarFallback>{getInitials(user().name)}</AvatarFallback>
            </Avatar>
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-56" align="end">
          <div className="flex flex-col space-y-1">
            <div className="px-2 py-1.5 text-sm font-medium flex items-center gap-2">
              <User className="h-4 w-4 ml-1 mr-2 text-muted-foreground" />
              {user().name}
            </div>
            <div className="h-px bg-border my-1" />
            <Button
              variant="ghost"
              className="justify-start"
              onClick={() => setIsProfileOpen(true)}
            >
              <UserCircle className="mr-2 h-4 w-4" />
              Profile
            </Button>
            <Button
              variant="ghost"
              className="justify-start"
              onClick={() => setIsPasswordOpen(true)}
            >
              <Key className="mr-2 h-4 w-4" />
              Password
            </Button>
            <Button
              variant="ghost"
              className="justify-start text-destructive hover:text-destructive"
              onClick={logout}
            >
              <LogOut className="mr-2 h-4 w-4" />
              Logout
            </Button>
          </div>
        </PopoverContent>
      </Popover>

      <ProfileModal isOpen={isProfileOpen} onOpenChange={setIsProfileOpen} />
      <PasswordModal isOpen={isPasswordOpen} onOpenChange={setIsPasswordOpen} />
    </>
  );
}
