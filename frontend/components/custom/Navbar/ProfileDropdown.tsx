"use client";

import { useLogout, useUser } from "@/components/hooks/useUser";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { FileText, Key, LogOut, User, UserCircle } from "lucide-react";
import { useState } from "react";

import { PasswordModal } from "../Modal/PasswordModal";
import { ProfileModal } from "../Modal/ProfileModal";
import { ViewStatementsModal } from "../Modal/Statement/ViewStatementsModal";

export function ProfileDropdown() {
  const { data: user } = useUser();
  const logoutMutation = useLogout();
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const [isPasswordOpen, setIsPasswordOpen] = useState(false);
  const [isStatementsOpen, setIsStatementsOpen] = useState(false);

  const getInitials = (name: string | null | undefined) => {
    if (!name) return "PN";
    return name
      .split(" ")
      .map((word) => word[0])
      .join("")
      .toUpperCase();
  };

  const handleLogout = () => {
    logoutMutation.mutate();
  };

  if (!user) {
    return (
      <Button variant="ghost" className="relative h-8 w-8 rounded-full">
        <Avatar className="h-8 w-8">
          <AvatarFallback>
            <UserCircle className="h-4 w-4" />
          </AvatarFallback>
        </Avatar>
      </Button>
    );
  }

  return (
    <>
      <Popover>
        <PopoverTrigger asChild>
          <Button variant="ghost" className="relative h-8 w-8 rounded-full">
            <Avatar className="h-8 w-8">
              <AvatarFallback>{getInitials(user.name)}</AvatarFallback>
            </Avatar>
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-56" align="end" forceMount>
          <div className="flex flex-col space-y-1">
            <p className="text-sm font-medium leading-none">{user.name}</p>
            <p className="text-xs leading-none text-muted-foreground">
              {user.email}
            </p>
          </div>
          <div className="mt-3 space-y-1">
            <Button
              variant="ghost"
              className="w-full justify-start"
              onClick={() => setIsProfileOpen(true)}
            >
              <User className="mr-2 h-4 w-4" />
              Profile
            </Button>
            <Button
              variant="ghost"
              className="w-full justify-start"
              onClick={() => setIsPasswordOpen(true)}
            >
              <Key className="mr-2 h-4 w-4" />
              Change Password
            </Button>
            <Button
              variant="ghost"
              className="w-full justify-start"
              onClick={() => setIsStatementsOpen(true)}
            >
              <FileText className="mr-2 h-4 w-4" />
              View Statements
            </Button>
            <Button
              variant="ghost"
              className="w-full justify-start"
              onClick={handleLogout}
              disabled={logoutMutation.isPending}
            >
              <LogOut className="mr-2 h-4 w-4" />
              Log out
            </Button>
          </div>
        </PopoverContent>
      </Popover>

      <ProfileModal isOpen={isProfileOpen} onOpenChange={setIsProfileOpen} />

      <PasswordModal isOpen={isPasswordOpen} onOpenChange={setIsPasswordOpen} />

      <ViewStatementsModal
        isOpen={isStatementsOpen}
        onOpenChange={setIsStatementsOpen}
      />
    </>
  );
}
